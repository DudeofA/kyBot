package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var s *discordgo.Session

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	log.Info("STARTING UP")

	ConnectDB()
	err := db.Migrator().AutoMigrate(&Server{}, &User{}, &ReadyCheck{}, &ReadyCheckStatus{})
	if err != nil {
		log.Fatalf("Database migration failed: %s", err.Error())
	}
	log.Infof("Connected to kyDB")

	s, err = discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatalln("Error creating Discord session :(", err)
	}

	s.Identify.Intents |= discordgo.IntentGuilds
	s.Identify.Intents |= discordgo.IntentGuildPresences
	s.Identify.Intents |= discordgo.IntentGuildMembers
	s.Identify.Intents |= discordgo.IntentGuildVoiceStates

	s.StateEnabled = true
	s.State.TrackChannels = true
	s.State.TrackMembers = true
	s.State.TrackVoice = true
	s.State.TrackPresences = true

	s.AddHandlerOnce(Ready)
	s.AddHandler(MessageCreate)
	s.AddHandler(ReactAdd)
	s.AddHandler(InteractionCreate)
	s.AddHandler(RateLimit)

	err = s.Open()
	if err != nil {
		log.Panicln("error opening discord connection :(", err)
	}
	defer s.Close()

	// Create channels to watch for kill signals
	botChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// Bot will end on any of the following signals
	signal.Notify(botChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		signalType := <-botChan
		log.Warningln("Shutting down from signal", signalType)
		done <- true
	}()

	// Wait here until CTRL-C or other term signal is received.
	<-done
}
