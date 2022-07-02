package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	*Globals

	TlgCmds map[string]func(opts *Opts)
	TxtCmds map[string]func(opts *Opts)

	*Media
	MainMenu tgbotapi.ReplyKeyboardMarkup
}

func NewBot(token string, debug bool) (*Bot, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
	}
	bot.Debug = debug

	myBot := &Bot{Globals: NewGlobals(bot, debug)}

	if err := myBot.initMedia(); err != nil {
		return nil, err
	}

	myBot.initMainMenu()

	return myBot, nil
}

type Globals struct {
	Bot   *tgbotapi.BotAPI
	Debug bool
}

func NewGlobals(bot *tgbotapi.BotAPI, debug bool) *Globals {
	return &Globals{Bot: bot, Debug: debug}
}

type Opts struct {
	Msg    *tgbotapi.Message
	NewMsg *tgbotapi.MessageConfig
}

func (b *Bot) registerUpdates(webHook, whPort, whPath string) (updates tgbotapi.UpdatesChannel, err error) {
	webHook = strings.TrimSpace(webHook)
	if webHook == "" {
		if _, err = b.Globals.Bot.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
			err = fmt.Errorf("bot.Request(tgbotapi.DeleteWebhookConfig{}): %w", err)
			return
		}
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates = b.Globals.Bot.GetUpdatesChan(u)
		return
	}

	// if webHook not empty run Telegram bot in web hook mode
	var wh tgbotapi.WebhookConfig
	wh, err = tgbotapi.NewWebhook(path.Join(webHook, whPath))
	if err != nil {
		err = fmt.Errorf("tgbotapi.NewWebhook() error: %w", err)
		return
	}
	if _, err = b.Globals.Bot.Request(wh); err != nil {
		err = fmt.Errorf("bot.Request webhook: %w", err)
		return
	}

	var info tgbotapi.WebhookInfo
	info, err = b.Globals.Bot.GetWebhookInfo()
	if err != nil {
		err = fmt.Errorf("bot.GetWebhookInfo(): %w", err)
		return
	}
	if info.LastErrorDate != 0 {
		err = fmt.Errorf("telegram callback failed: %s", info.LastErrorMessage)
		return
	}
	go http.ListenAndServe(whPort, nil)
	updates = b.Globals.Bot.ListenForWebhook(whPath)
	return
}

func (b *Bot) initMedia() (err error) {
	b.Media, err = NewMedia()
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

func (b *Bot) printCommands() {
	fmt.Println("show commands")
	for _, cat := range b.Media.Categories {
		fmt.Println(cat.Name, "- random", cat.TxtName)
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) (err error) {
	newMsg := tgbotapi.NewMessage(msg.Chat.ID, "")
	defer func() {
		if newMsg.Text != "" {
			_, err = b.Bot.Send(newMsg)
		}
	}()

	msg.Text = strings.TrimSpace(msg.Text)

	if msg.IsCommand() {
		// process media
		for _, cat := range b.Media.Categories {
			if cat.Name == msg.Command() {
				if err := b.SendMediaToChat(msg.Chat.ID, cat.ID); err != nil {
					log.Println(err)
				}
				return nil
			}
		}
		//botcmd.HandleCmd(msg, &newMsg)
		return nil
	}

	return nil
}

func (b *Bot) SendMediaToChat(chatID int64, imgCatID uint8) error {
	img := b.Media.randomImage(imgCatID)
	ext := strings.ToLower(filepath.Ext(img))

	var myMsg tgbotapi.Chattable
	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", ".webp":
		myMsg = tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(img))
	case ".gif", ".mp4":
		/*myA := tgbotapi.NewAnimation(chatID, tgbotapi.FileURL(img))
		myA.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ch","ch")))
		myMsg = myA*/
		myMsg = tgbotapi.NewAnimation(chatID, tgbotapi.FileURL(img))
	/*case ".webm":
	myMsg = tgbotapi.VideoConfig{}*/
	default:
		myMsg = tgbotapi.NewDocument(chatID, tgbotapi.FileURL(img))
	}
	_, err := b.Bot.Send(myMsg)
	return err
}
