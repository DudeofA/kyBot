package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
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
	MigrateDB()
	err := db.Migrator().AutoMigrate(&Server{}, &Wordle{}, &WordleStat{}, &User{})
	if err != nil {
		log.Fatalf("Database migration failed: %s", err.Error())
	}
	log.Infof("Connected to kyDB")

	// Session
	s, err = discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatalln("Error creating Discord session :(", err)
	}
	defer s.Close()

	s.AddHandlerOnce(Ready)
	s.AddHandler(MessageCreate)
	s.AddHandler(ReactAdd)
	s.AddHandler(InteractionCreate)
	s.AddHandler(RateLimit)

	err = s.Open()
	if err != nil {
		log.Panicln("error opening discord connection :(", err)
	}

	c := cron.New()
	log.Debug("Wordle reminders will go out each day at 12am")
	c.AddFunc("0 0 0 * * *", func() { SendWordleReminders() })
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
