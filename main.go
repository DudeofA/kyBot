package main

import (
	"os"
	"os/signal"
	"syscall"

	"kyBot/component"
	"kyBot/config"
	"kyBot/handlers"
	"kyBot/kyDB"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	log.Info("STARTING UP")

	kyDB.Connect()
	err := kyDB.DB.Migrator().AutoMigrate(&component.Server{}, &component.Wordle{}, &component.WordleStat{}, &component.User{})
	if err != nil {
		log.Fatalf("Database migration failed: %s", err.Error())
	}
	log.Infof("Connected to kyDB")

	// Session
	s, err := discordgo.New("Bot " + config.TOKEN)
	if err != nil {
		log.Fatalln("Error creating Discord session :(", err)
	}
	defer s.Close()

	s.AddHandlerOnce(handlers.Ready)
	s.AddHandler(handlers.MessageCreate)
	s.AddHandler(handlers.ReactAdd)
	s.AddHandler(handlers.InteractionCreate)
	s.AddHandler(handlers.RateLimit)

	err = s.Open()
	if err != nil {
		log.Panicln("Error openning connection :(", err)
	}

	c := cron.New()
	log.Debug("Wordle reminders will go out each day at 12am")
	c.AddFunc("0 0 0 * * *", func() { component.SendWordleReminders(s) })
	c.Start()

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
