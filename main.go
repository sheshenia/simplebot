package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func run(ctx context.Context) error {
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
			log.Println(update)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Program halted!")
		signal.Stop(c)
		cancel()
	}()
	if err := run(ctx); err != nil {
		log.Fatalln(err)
	}
}
