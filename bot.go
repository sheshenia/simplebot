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

	// TlgCmds bot commands
	TlgCmds map[string]func(opts *Opts)
	// Text commands, from buttons or text
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
	myBot.initTlgCmds()
	myBot.initTextCmds()

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

func (b *Bot) handleMessage(msg *tgbotapi.Message) (err error) {
	newMsg := tgbotapi.NewMessage(msg.Chat.ID, "")
	msg.Text = strings.TrimSpace(msg.Text)
	opts := Opts{
		Msg:    msg,
		NewMsg: &newMsg,
	}

	defer func() {
		if opts.NewMsg.Text != "" {
			_, err = b.Bot.Send(opts.NewMsg)
		}
	}()

	if msg.IsCommand() {
		if b.handleMediaTlgTxtCommand(&opts, true) {
			return nil
		}
		return b.handleBotCommand(&opts)
	}

	if b.handleTextCommand(&opts) {
		return nil
	}

	// process media buttons click keyboard
	if b.handleMediaTlgTxtCommand(&opts, false) {
		return nil
	}

	opts.NewMsg.Text = TextUnknownMessage
	opts.NewMsg.ReplyMarkup = b.MainMenu
	return nil
}

func (b *Bot) handleBotCommand(opts *Opts) error {
	// if founded, process command
	if cmd, ok := b.TlgCmds[opts.Msg.Command()]; ok {
		cmd(opts)
		return nil
	}
	opts.NewMsg.Text = ErrUnknownCommand.Error()
	opts.NewMsg.ReplyMarkup = b.MainMenu
	return ErrUnknownCommand
}

func (b *Bot) handleTextCommand(opts *Opts) bool {
	// if founded, process text command
	if cmd, ok := b.TxtCmds[opts.Msg.Text]; ok {
		cmd(opts)
		return true
	}
	return false
}

func (b *Bot) handleMediaTlgTxtCommand(opts *Opts, isCmd bool) bool {
	for _, cat := range b.Media.Categories {
		if (isCmd && cat.Name == opts.Msg.Command()) || (!isCmd && cat.TxtName == opts.Msg.Text) {
			if err := b.SendMediaToChat(opts.Msg.Chat.ID, cat.ID); err != nil {
				log.Println(err)
			}
			return true
		}
	}
	return false
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
