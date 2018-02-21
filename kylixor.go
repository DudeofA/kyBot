package main

import (
	//	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
    "github.com/bwmarrin/dgvoice"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string
var logID = "296765401944162305"

func main() {

	if token == "" {
		fmt.Println("No token provided. Please run: kylixor -t <bot token>")
		return
	}

	// Create a new Discord session using the provided bot token.
	ky, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	ky.AddHandlerOnce(ready)

	// Register messageCreate as a callback for the messageCreate events.
	ky.AddHandler(messageCreate)

	// Register other things
	// ky.AddHandler(messageReactionAdd)
	// ky.AddHandler(messageReactionRemove)

	// Open the websocket and begin listening.
	err = ky.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Kylixor is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	ky.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateStatus(0, "Alpha beta early access")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

    // Log every message in log channel
	if m.ChannelID != logID {
		timestamp := time.Now()
		logMessage(s, timestamp, m.Message.Author, m.ID, m.ChannelID, "MSG", m.ContentWithMentionsReplaced())
	}

    // Testing by me only
    if m.Content == "test" && m.Author.ID == "144220178853396480" {
        Test(s, m)
    }

    // Plays the 'yee' clip that is so close to our hearts
    if m.Content == "yee" {
        // Apparently I need this
        stopChan := make(chan bool)
        c, _ := s.State.Channel(m.ChannelID)
        g, _ := s.State.Guild(c.GuildID)
        // Search through the guild's voice channels for the command's author
        for _, vs := range g.VoiceStates {
            if vs.UserID == m.Author.ID {
                voiceChannel, _ := s.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, false);
                // TO REPLACE WITH MY OWN CODE
                dgvoice.PlayAudioFile(voiceChannel, "clips/yee.mp3", stopChan);
                voiceChannel.Disconnect();
            }
        }
    }

    if m.Content == "help" {
        readme, err := ioutil.ReadFile("README.md")
        if err != nil {
            panic(err)
        }
        // Print readme in code brackets so it doesn't look awful
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```" + string(readme) + "```"))
    }
	if m.Content == "ping" {
        pongMessage, _ := s.ChannelMessageSend(m.ChannelID, "Pong!")
        pongStamp, _ := time.Parse("2006-01-02T15:04:05-07:00", string(pongMessage.Timestamp))
        duration := time.Since(pongStamp)
        pingTime := duration.Nanoseconds() / 1000000
        s.ChannelMessageEdit(m.ChannelID, pongMessage.ID, fmt.Sprintf("Pong! %vms", pingTime))

	}

    if m.Content == "pizza" {
        s.ChannelMessageSend(m.ChannelID, "🍕 here it is, come get it. \nI ain't your delivery bitch.")
    }

	if strings.HasPrefix(strings.ToLower(m.Content), "quote ") {
        //NEEDS IMPROVEMENTS
        quote := strings.TrimPrefix(m.Content, "quote ")
        quote = strings.TrimPrefix(quote, "Quote ")
        Vote(s, m, quote)
	}

    if strings.HasPrefix(strings.ToLower(m.Content), "quotelist"){
        entries := ListQuote(s, m)
        if entries <= 1 {
            s.ChannelMessageSend(m.ChannelID, "No quotes in file")
        }
    }
    //only Me
    if m.Content == "quoteclear" && m.Author.ID == "144220178853396480" {
        _, _ = os.Create("quotes.txt")
        s.ChannelMessageSend(m.ChannelID, "Quote file cleared")
    }

    if m.Content == "randquote" {
        ShowRandQuote(s, m)
    }
}
