package bot

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sheshenia/simplebot/pkg/caption"
	"github.com/sheshenia/simplebot/pkg/media"
)

// Bot is our main structure with all info and methods to process (handle)
// bot commands, text buttons commands and other
type Bot struct {
	// Telegram bot API, debug end other general data
	*Globals

	// Telegram bot commands /start /stop /home etc...
	TlgCmds map[string]func(opts *Opts)

	// Text commands, from buttons or text
	TxtCmds map[string]func(opts *Opts)

	// media categories' info, containing all we need to handle media (commands, buttons pressed etc)
	*media.Media

	// Main buttons menu, containing media categories data
	MainMenu tgbotapi.ReplyKeyboardMarkup
}

// NewBot creates bot instance. Init all data like commands, text commands, media categories
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

// Opts is used in bot & txt command function handlers
type Opts struct {
	Msg    *tgbotapi.Message
	NewMsg *tgbotapi.MessageConfig
}

// registerUpdates creates subscriptions to the updates channel. In common or webhook mode.
// What to choose depends on bot usage. But in general if you plan to use on server
// with multiple users Webhook is preferred.
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

// handleMessage processes update.Message from Telegram API.
// It can be bot commands or text commands from buttons.
// Here we detect all possible scenarios and handle them.
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

	// process media buttons click from MainMenu
	if b.handleMediaTlgTxtCommand(&opts, false) {
		return nil
	}

	// if no handlers founded inform user
	opts.NewMsg.Text = caption.TextUnknownMessage
	opts.NewMsg.ReplyMarkup = b.MainMenu
	return nil
}

func (b *Bot) handleBotCommand(opts *Opts) error {
	// if founded, process command
	if cmd, ok := b.TlgCmds[opts.Msg.Command()]; ok {
		cmd(opts)
		return nil
	}
	opts.NewMsg.Text = caption.ErrUnknownCommand.Error()
	opts.NewMsg.ReplyMarkup = b.MainMenu
	return caption.ErrUnknownCommand
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
	img := b.Media.RandomImage(imgCatID)
	ext := strings.ToLower(filepath.Ext(img))

	var (
		myMsg  tgbotapi.Chattable
		imgRfd tgbotapi.RequestFileData
	)
	if strings.HasPrefix(img, "http") {
		imgRfd = tgbotapi.FileURL(img)
	} else {
		imgRfd = tgbotapi.FilePath(img)
	}
	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", ".webp":
		myMsg = tgbotapi.NewPhoto(chatID, imgRfd)
	case ".gif", ".mp4":
		/*myA := tgbotapi.NewAnimation(chatID, tgbotapi.FileURL(img))
		myA.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ch","ch")))
		myMsg = myA*/
		myMsg = tgbotapi.NewAnimation(chatID, imgRfd)
	/*case ".webm":
	myMsg = tgbotapi.VideoConfig{}*/
	default:
		myMsg = tgbotapi.NewDocument(chatID, imgRfd)
	}
	_, err := b.Bot.Send(myMsg)
	return err
}
