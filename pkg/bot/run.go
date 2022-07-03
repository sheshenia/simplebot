package bot

import (
	"context"
	"flag"
	"log"
)

func Run(ctx context.Context) error {
	var (
		token        = flag.String("token", "", "telegram bot  token")
		debug        = flag.Bool("debug", false, "debug mode")
		webHook      = flag.String("webhook", "", "if not empty Telegram bot starts in WebHook mode. ex: examplebot.com")
		whPort       = flag.String("port", ":8081", "web hook port")
		whPath       = flag.String("path", "/", "web hook path")
		showCommands = flag.Bool("commands", false, "show bot commands at startup")
	)
	flag.Parse()

	bot, err := NewBot(*token, *debug)
	if err != nil {
		return err
	}

	if *showCommands {
		bot.printCommands()
	}

	updates, err := bot.registerUpdates(*webHook, *whPort, *whPath)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			if update.Message != nil {
				if err := bot.handleMessage(update.Message); err != nil && bot.Debug {
					log.Println(err)
				}
			}
		}
	}
}
