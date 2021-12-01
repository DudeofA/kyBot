package main

import (
	"bytes"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {

	var (
		buf  bytes.Buffer
		klog = log.New(&buf, "kyBot: ", log.Lshortfile|log.Ldate|log.Ltime)
	)

	godotenv.Load()

	klog.SetOutput(os.Stdout)

	klog.Println("STARTING UP")

	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		klog.Fatal("Error creating Discord session :(", err)
		return
	}
	defer s.Close()

	err = s.Open()
	if err != nil {
		klog.Fatal("Error openning connection :(", err)
		return
	}

	// Create channels to watch for kill signals
	botChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// Bot will end on any of the following signals
	signal.Notify(botChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		signalType := <-botChan
		klog.Println("Shutting down from signal", signalType)
		done <- true
	}()

	// Wait here until CTRL-C or other term signal is received.
	<-done

}
