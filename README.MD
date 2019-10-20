# Password Generator Bot
A simple Telegram bot to generate secure random passwords
## Features 
* **Lightweight**: 200 lines of code, nearly low memory usage.
* **Secure**: Use `crypto/rand` to generate secure random passwords.
* **No trace**: Nothing is stored on server and no logs.
## Live Bot
I try to keep this bot alive.

http://t.me/pass_generator_bot
## Install
### Building
```bash
git clone https://github.com/HirbodBehnam/PasswordGeneratorBot
cd PasswordGeneratorBot
go get github.com/go-telegram-bot-api/telegram-bot-api
go build main.go
```
### Running
```bash
./main <BOT_TOKEN>
```