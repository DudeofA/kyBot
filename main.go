package main

import (
	"os"
	"os/signal"
	"syscall"

	"kyBot/handlers"
	"kyBot/kyDB"
	"kyBot/minecraft"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

var (
	s *discordgo.Session
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	godotenv.Load()
	token, token_found := os.LookupEnv("DISCORD_TOKEN")
	if !token_found {
		log.Fatal("No token found, please set env DISCORD_TOKEN to a valid Discord bot token")
	}

	log.Info("STARTING UP")

	db := kyDB.Connect()
	db.AutoMigrate(&kyDB.User{}, &kyDB.Guild{}, &kyDB.Hangman{}, &minecraft.MinecraftServer{})

	var err error
	s, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Error creating Discord session :(", err)
	}
	defer s.Close()

	s.AddHandlerOnce(handlers.Ready)
	s.AddHandler(handlers.MessageCreate)
	s.AddHandler(handlers.ReactAdd)

	err = s.Open()
	if err != nil {
		log.Panicln("Error openning connection :(", err)
	}

	c := cron.New()
	// log.Info("Updating Minecraft servers every minute")
	// c.AddFunc("0 * * * * *", func() { minecraft.UpdateAllServers(s) })
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
