package main

import (
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	*Globals

	TlgCmds map[string]func(opts *Opts)
	TxtCmds map[string]func(opts *Opts)
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

	globals := NewGlobals(bot, debug)

	return &Bot{Globals: globals}, nil
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
