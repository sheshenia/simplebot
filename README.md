# SimpleBot
Simple Telegram images bot - works out of the box.

A good starting point for your Telegram bots and Go beginners.

SimpleBot is written with Golang programming language and uses the [Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api) package under the hood.
![gif](overview.gif)

The step-by-step tutorial of creating this bot will be available soon.

## The quick-start guide on a local machine:
[Download release archive and unpack](https://github.com/sheshenia/simplebot/releases)
* Linux: simplebot-linux-amd64-*.zip
* Windows: simplebot-windows-amd64-*.zip
* Mac intel: simplebot-darwin-amd64-*.zip
* Mac M-1, M-2: simplebot-darwin-arm64-*.zip

Register your bot in Telegram and get your bot token
`9876543210:SomeDummyTokenFromBotFatherTelegram`

### Starting on Linux and Mac
* Navigate to the unpacked bot folder, and start
```
./simplebot-linux-amd64 -token 9876543210:SomeDummyTokenFromBotFatherTelegram -debug
```
* This command will start locally your telegram bot in debug mode
* Type in Telegram the name of your bot and enjoy using it!

### Starting on Windows (tutorial soon)
* Open your command line and type
```
./simplebot-windows-amd64.exe -token 9876543210:SomeDummyTokenFromBotFatherTelegram -debug
```
*  Type in Telegram the name of your bot and enjoy using it!

### Deploy on server in webhook mode
* Prepare your Telegram bot to be used in webhook mode in Telegram
* Unpack simplebot-linux-amd64-*.zip on server
* Prepare NGINX (Apache) to be used with SimpleBot as proxy servers
* Prepare Nginx to work serve https, and correctly local proxy to the SimpleBot port. 
* Navigate to the SimpleBot folder and start your bot
```
./simplebot-linux-amd64 \
-token 9876543210:SomeDummyTokenFromBotFatherTelegram \
-webhook example.com
```
* This command will start your bot in webhook mode on https://example.com/ address
* Use `systemd` or `supervisorctl` service to start you bot on system start-up.

### Start on server in webhook mode with custom port and path
```
./simplebot-linux-amd64 \
-token 9876543210:SomeDummyTokenFromBotFatherTelegram \
-webhook example.com \
-path /some_path \
-port :8085
```
This command will start your bot in webhook mode on custom address https://example.com/some_path

On custom port :8085

