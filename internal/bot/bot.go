package bot

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/joehonkey/msg2irc/internal/config"
)

type Bot struct {
	conn     *ircevent.Connection
	channels map[string]bool
	mu       sync.Mutex
}

func New(cfg *config.Config) (*Bot, error) {
	tlsHostname := cfg.IRC.TLSHostname
	if tlsHostname == "" {
		tlsHostname = cfg.IRC.Server
	}

	conn := &ircevent.Connection{
		Server:      fmt.Sprintf("%s:%d", cfg.IRC.Server, cfg.IRC.Port),
		Nick:        cfg.IRC.Nick,
		User:        cfg.IRC.Nick,
		UseTLS:      cfg.IRC.TLS,
		RequestCaps: []string{"message-tags"},
	}
	if cfg.IRC.TLS {
		conn.TLSConfig = &tls.Config{ServerName: tlsHostname}
	}

	b := &Bot{
		conn:     conn,
		channels: make(map[string]bool),
	}

	conn.AddCallback("001", func(msg ircmsg.Message) {
		for _, ch := range cfg.IRC.Channels {
			log.Printf("joining %s", ch)
			conn.Join(ch)
			b.mu.Lock()
			b.channels[ch] = true
			b.mu.Unlock()
		}
	})

	return b, nil
}

func (b *Bot) Run() {
	if err := b.conn.Connect(); err != nil {
		log.Fatalf("irc connect: %v", err)
	}
	b.conn.Loop()
}

func (b *Bot) Send(target, message string) error {
	if !b.conn.Connected() {
		return fmt.Errorf("not connected to IRC")
	}

	if strings.HasPrefix(target, "#") {
		b.mu.Lock()
		if !b.channels[target] {
			b.conn.Join(target)
			b.channels[target] = true
		}
		b.mu.Unlock()
	}

	return b.conn.Privmsg(target, message)
}
