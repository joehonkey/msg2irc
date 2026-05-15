package server

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"

	"github.com/joehonkey/msg2irc/internal/bot"
	"github.com/joehonkey/msg2irc/internal/config"
)

type Server struct {
	cfg *config.Config
	bot *bot.Bot
}

type msgRequest struct {
	Target  string `json:"target"`
	Message string `json:"message"`
	From    string `json:"from"`
	Token   string `json:"token"`
}

func New(cfg *config.Config, b *bot.Bot) *Server {
	return &Server{cfg: cfg, bot: b}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/send", s.handleSend)

	switch s.cfg.TLS.Mode {
	case "auto":
		return s.startAuto(mux)
	case "manual":
		return s.startManual(mux)
	default: // off — plain HTTP, meant to sit behind a reverse proxy
		log.Printf("TLS off — listening plain HTTP on %s", s.cfg.ListenAddr)
		return http.ListenAndServe(s.cfg.ListenAddr, mux)
	}
}

func (s *Server) startAuto(mux *http.ServeMux) error {
	cacheDir := s.cfg.TLS.CacheDir
	if cacheDir == "" {
		cacheDir = "./certs"
	}
	listenAddr := s.cfg.TLS.ListenAddr
	if listenAddr == "" {
		listenAddr = ":443"
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.cfg.TLS.Domain),
		Cache:      autocert.DirCache(cacheDir),
	}

	// HTTP on :80 handles Let's Encrypt challenges + redirects to HTTPS
	go func() {
		log.Println("HTTP redirect listening on :80")
		if err := http.ListenAndServe(":80", m.HTTPHandler(nil)); err != nil {
			log.Printf("HTTP redirect: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:      listenAddr,
		Handler:   mux,
		TLSConfig: m.TLSConfig(),
	}
	log.Printf("TLS auto — listening on %s for %s", listenAddr, s.cfg.TLS.Domain)
	return srv.ListenAndServeTLS("", "")
}

func (s *Server) startManual(mux *http.ServeMux) error {
	listenAddr := s.cfg.TLS.ListenAddr
	if listenAddr == "" {
		listenAddr = ":443"
	}
	log.Printf("TLS manual — listening on %s", listenAddr)
	return http.ListenAndServeTLS(listenAddr, s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile, mux)
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req msgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Token != s.cfg.Token {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if req.Target == "" || req.Message == "" {
		http.Error(w, "target and message required", http.StatusBadRequest)
		return
	}

	msg := req.Message
	if req.From != "" {
		msg = req.From + ": " + req.Message
	}

	if err := s.bot.Send(req.Target, msg); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}
