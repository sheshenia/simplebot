package main

import (
	"errors"
	"fmt"
	"net/http"
	"path"
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

	if len(myBot.Media.IDs) == 0 {
		return nil, errors.New("no media categories")
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
	for key, id := range b.Media.IDs {
		if (key)%3 == 0 {
			b.MainMenu.Keyboard = append(b.MainMenu.Keyboard, tgbotapi.NewKeyboardButtonRow())
		}
		btn := tgbotapi.NewKeyboardButton(b.Media.Categories[id].TxtName)
		r := len(b.MainMenu.Keyboard) - 1 // last row ID
		b.MainMenu.Keyboard[r] = append(b.MainMenu.Keyboard[r], btn)
	}
}
