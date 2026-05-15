package main

import (
	"flag"
	"log"

	"github.com/joehonkey/msg2irc/internal/bot"
	"github.com/joehonkey/msg2irc/internal/config"
	"github.com/joehonkey/msg2irc/internal/server"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}

	go b.Run()

	s := server.New(cfg, b)
	log.Printf("msg2irc listening on %s", cfg.ListenAddr)
	log.Fatal(s.Start())
}
