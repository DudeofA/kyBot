package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func RateLimit(s *discordgo.Session, r *discordgo.RateLimit) {
	log.Warnf("Rate limit hit: %s", r.Message)
	s.Ratelimiter.Mutex.Lock()
	time.Sleep(r.RetryAfter)
	s.Ratelimiter.Mutex.Unlock()
}
