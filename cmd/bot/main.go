package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sheshenia/simplebot/pkg/bot"
)

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
	if err := bot.Run(ctx); err != nil {
		log.Fatalln(err)
	}
}
