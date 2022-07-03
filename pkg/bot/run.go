package bot

import (
	"context"
	"flag"
	"fmt"
	"log"
)

var (
	version = "unknown" // overridden by -ldflags -X
)

func Run(ctx context.Context) error {
	var (
		flToken        = flag.String("token", "", "telegram bot  token")
		flDebug        = flag.Bool("debug", false, "debug mode")
		flWebHook      = flag.String("webhook", "", "if not empty Telegram bot starts in WebHook mode. ex: examplebot.com")
		flWhPort       = flag.String("port", ":8081", "web hook port")
		flWhPath       = flag.String("path", "/", "web hook path")
		flShowCommands = flag.Bool("commands", false, "show bot commands at startup")
		flVersion      = flag.Bool("version", false, "print version")
	)
	flag.Parse()

	if *flVersion {
		fmt.Println(version)
		return nil
	}

	bot, err := NewBot(*flToken, *flDebug)
	if err != nil {
		return err
	}

	if *flShowCommands {
		bot.printCommands()
	}

	updates, err := bot.registerUpdates(*flWebHook, *flWhPort, *flWhPath)
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
