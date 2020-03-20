package main

import (
	"crypto/rand"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/patrickmn/go-cache"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

type UsersIn struct {
	PageIn    byte
	Length    byte
	LowerCase bool
	UpperCase bool
	Numbers   bool
	Special   bool
}

var users *cache.Cache

const LettersLower = "qwertyuiopasdfghjklzxcvbnm"
const LettersUpper = "QWERTYUIOPASDFGHJKLZXCVBNM"
const NUMBERS = "1234567890"
const SYMBOLS = `!@#$%^&*()_+=-[]{};:'"\|,./~`

const VERSION = "1.0.0"

var ynKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Yes"),
		tgbotapi.NewKeyboardButton("No"),
	),
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please pass the bot token as argument.")
	}
	//Load bot
	bot, err := tgbotapi.NewBotAPI(os.Args[1])
	if err != nil {
		panic("Cannot initialize the bot: " + err.Error())
	}
	log.Println("Password Generator Bot v" + VERSION)
	log.Println("Bot authorized on account", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	users = cache.New(5*time.Minute, 10*time.Minute)

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message
			continue
		}
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Hello and welcome to password generator bot!\nTo quickly generate a password send run /generate\nTo customize your password use /password"))
			case "about":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Password Generator Bot v"+VERSION+"\nBy Hirbod Behnam\nSource: https://github.com/HirbodBehnam/PasswordGeneratorBot"))
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "This bot helps you generate random passwords. **I DO NOT STORE ANYTHING ON MY SERVER**, you can read the source code. Also this bot uses `crypto/rand` to generate secure randoms for password.\nTo quickly generate password use /generate , it generates a 16 letter password with combination of letters and numbers\nIf you want to create a customizable password, use /password")
				msg.ParseMode = "markdown"
				_, _ = bot.Send(msg)
			case "generate":
				go func(id int64) {
					//Generate a 16 length password with all alphabet and numbers
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "`"+GeneratePassword(16, true, true, true, false)+"`")
					msg.ParseMode = "markdown"
					_, _ = bot.Send(msg)
				}(update.Message.Chat.ID)
			case "password":
				users.Set(strconv.Itoa(update.Message.From.ID), &UsersIn{PageIn: 0, Length: 0, LowerCase: false, UpperCase: false, Numbers: false, Special: false}, cache.DefaultExpiration)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select the length of your password(1-255)")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				_, _ = bot.Send(msg)
			default:
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry this command is not recognized; Try /help"))
			}
		} else {
			if userT, ok := users.Get(strconv.Itoa(update.Message.From.ID)); ok {
				user := userT.(*UsersIn)
				switch user.PageIn {
				case 0: //User should have entered the length of password
					//Try to parse the password
					num, err := strconv.Atoi(update.Message.Text)
					if err != nil {
						_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Cannot parse the value you send: "+err.Error()))
						continue
					}
					if num > 255 || num < 1 {
						_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter a value less than 255 and more than zero"))
						continue
					}
					user.Length = byte(num)
					user.PageIn++
					//Inform user
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Good. Do you want your password contain lower case characters? (a,b,c...)")
					msg.ReplyMarkup = ynKeyboard
					_, _ = bot.Send(msg)
				case 1: //Should password contain lowercase values?
					user.LowerCase = update.Message.Text == "Yes"
					user.PageIn++
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Do you want your password contain upper case characters? (A,B,C...)")
					msg.ReplyMarkup = ynKeyboard
					_, _ = bot.Send(msg)
				case 2: //Should password contain uppercase values?
					user.UpperCase = update.Message.Text == "Yes"
					user.PageIn++
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Do you want your password contain numbers characters? (1,2,3...)")
					msg.ReplyMarkup = ynKeyboard
					_, _ = bot.Send(msg)
				case 3: //Should password contain numbers?
					user.Numbers = update.Message.Text == "Yes"
					user.PageIn++
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Do you want your password contain special characters? (!,#,%...)")
					msg.ReplyMarkup = ynKeyboard
					_, _ = bot.Send(msg)
				case 4: //Should password contain symbols?
					user.Special = update.Message.Text == "Yes"
					//Generate password
					go func(in UsersIn, chatID int64) {
						msg := tgbotapi.NewMessage(chatID, "")
						msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						if !in.LowerCase && !in.UpperCase && !in.Numbers && !in.Special {
							msg.Text = "Please at least specify one character type"
							_, _ = bot.Send(msg)
							return
						}
						msg.Text = "```\n" + GeneratePassword(int(in.Length), in.LowerCase, in.UpperCase, in.Numbers, in.Special) + "\n```"
						msg.ParseMode = "markdown"
						_, err := bot.Send(msg)
						if err != nil {
							msg.Text = "Error sending password: " + err.Error()
							_, _ = bot.Send(msg)
						}
					}(*user, update.Message.Chat.ID)
					users.Delete(strconv.Itoa(update.Message.From.ID))
				}
			} else {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry this command is not recognized; Try /help"))
			}
		}
	}
}

func GeneratePassword(length int, lowercase, uppercase, number, symbol bool) string {
	str := ""
	if lowercase {
		str += LettersLower
	}
	if uppercase {
		str += LettersUpper
	}
	if number {
		str += NUMBERS
	}
	if symbol {
		str += SYMBOLS
	}
	//Memory shit
	var sb strings.Builder
	max := big.NewInt(int64(len(str)))
	sb.Grow(length)
	for i := 0; i < length; i++ {
		index, _ := rand.Int(rand.Reader, max)
		sb.WriteByte(str[index.Int64()])
	}
	return sb.String()
}
