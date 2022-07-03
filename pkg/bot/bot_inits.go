package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sheshenia/simplebot/pkg/caption"
	"github.com/sheshenia/simplebot/pkg/media"
)

func (b *Bot) initMedia() (err error) {
	b.Media, err = media.New()
	return
}

func (b *Bot) initMainMenu() {
	b.MainMenu = tgbotapi.NewReplyKeyboard()
	for key, cat := range b.Media.Categories {
		if (key)%3 == 0 {
			b.MainMenu.Keyboard = append(b.MainMenu.Keyboard, tgbotapi.NewKeyboardButtonRow())
		}
		btn := tgbotapi.NewKeyboardButton(cat.TxtName)
		r := len(b.MainMenu.Keyboard) - 1 // last row ID
		b.MainMenu.Keyboard[r] = append(b.MainMenu.Keyboard[r], btn)
	}
}

// initTlgCmds init bot commands (note: media commands in media.Categories -> Name):
// /start
// /stop
// ...
func (b *Bot) initTlgCmds() {
	b.TlgCmds = map[string]func(p *Opts){
		caption.CmdStart: b.DefaultStartCommand,
		caption.CmdStop: func(p *Opts) {
			//TODO here we should do all we need after bot stopped
			p.NewMsg.Text = "All tasks are deleted!"
			p.NewMsg.ParseMode = "HTML"
			p.NewMsg.ReplyMarkup = b.MainMenu
		},
		caption.CmdHome: b.DefaultStartCommand,
	}
}

// initTextCmds init bot text commands from buttons or input text
func (b *Bot) initTextCmds() {
	b.TxtCmds = map[string]func(p *Opts){
		caption.CmdTMainMenu: b.DefaultStartCommand,
	}
}

func (b *Bot) DefaultStartCommand(p *Opts) {
	//p.NewMsg.ParseMode = "HTML"
	p.NewMsg.Text = caption.TextCmdStart
	p.NewMsg.ReplyMarkup = b.MainMenu
}

func (b *Bot) printCommands() {
	fmt.Println("show commands")
	for _, cat := range b.Media.Categories {
		fmt.Println(cat.Name, "- random", cat.TxtName)
	}
}
