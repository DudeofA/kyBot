package main

import (
	"os"
	"os/signal"
	"syscall"

	"kyBot/handlers"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	godotenv.Load()

	logrus.Info("STARTING UP")

	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		logrus.Fatal("Error creating Discord session :(", err)
		return
	}
	defer s.Close()

	s.AddHandlerOnce(handlers.Ready)

	err = s.Open()
	if err != nil {
		logrus.Fatal("Error openning connection :(", err)

		return
	}

	// Create channels to watch for kill signals
	botChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// Bot will end on any of the following signals
	signal.Notify(botChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		signalType := <-botChan
		logrus.Info("Shutting down from signal: ", signalType)
		done <- true
	}()

	// Wait here until CTRL-C or other term signal is received.
	<-done

}
