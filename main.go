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
		token = flag.String("token", "", "telegram bot  token")
		debug = flag.Bool("debug", false, "debug mode")
	)
	flag.Parse()

	bot, err := NewBot(*token, *debug)
	if err != nil {
		return err
	}

	log.Println(bot)

	return nil
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
