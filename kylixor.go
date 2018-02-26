package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
    "encoding/json"

	"github.com/bwmarrin/discordgo"
    "github.com/bwmarrin/dgvoice"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string
var stopChan = make(chan bool)

type Config struct {
    Admin   string
    Yee     bool
    LogID   string
    Test    []string
}

var config = Config{}

func ReadConfig() {
    file, _ := os.Open("conf.json")
    decoder := json.NewDecoder(file)
    config = Config{}
    err := decoder.Decode(&config)
    if err != nil {
        fmt.Println("error: ", err)
    }
}

func main() {
    ReadConfig()

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
	s.UpdateStatus(0, "Alpha v0.2")
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
	if m.ChannelID != config.LogID {
		timestamp := time.Now()
		logMessage(s, timestamp, m.Message.Author, m.ID, m.ChannelID, "MSG", m.ContentWithMentionsReplaced())
	}

    if strings.HasPrefix(m.Content, "!") {
        // Remove prefix for 'performance'
        input := strings.TrimPrefix(m.Content, "!")

        switch input {

        case "help":
            readme, err := ioutil.ReadFile("README.md")
            if err != nil { panic(err) }
            // Print readme in code brackets so it doesn't look awful
            s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```" + string(readme) + "```"))
            break

        case "ping":
            pongMessage, _ := s.ChannelMessageSend(m.ChannelID, "Pong!")
            // Format Discord time to readable time
            pongStamp, _ := time.Parse("2006-01-02T15:04:05-07:00", string(pongMessage.Timestamp))
            duration := time.Since(pongStamp)
            pingTime := duration.Nanoseconds() / 1000000
            // Print duration from message being send to message being posted
            s.ChannelMessageEdit(m.ChannelID, pongMessage.ID, fmt.Sprintf("Pong! %vms", pingTime))
            break


        case "pizza":
            s.ChannelMessageSend(m.ChannelID, "🍕 here it is, come get it. \nI ain't your delivery bitch.")
            break

        case "quoteclear":
            if m.Author.ID == config.Admin {
                // Create empty file to overwrite old quote DANGER: CAN'T UNDO
                _, _ = os.Create("quotes.txt")
                s.ChannelMessageSend(m.ChannelID, "Quote file cleared")
            }
            break

        case "quotelist":
            entries := ListQuote(s, m)
            if entries <= 1 {
                s.ChannelMessageSend(m.ChannelID, "No quotes in file")
            }
            break

        case "randquote":
            ShowRandQuote(s, m)
            break

        case "reload":
            if m.Author.ID == config.Admin {
                ReadConfig()
                s.ChannelMessageSend(m.ChannelID, "Config reloaded")
            }
            break

        case "test":
            if m.Author.ID == config.Admin {
                Test(s,m)
            }
            break

        // Plays the 'yee' clip that is so close to our hearts
        case "yee":
            if !config.Yee {
                s.ChannelMessageSend(m.ChannelID, "Yee is disabled")
            } else {
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

        default:
        //    s.ChannelMessageSend(m.ChannelID, "Whatchu say?")
        }

        if strings.HasPrefix(strings.ToLower(m.Content), "quote ") {
            //NEEDS IMPROVEMENTS
            quote := strings.TrimPrefix(m.Content, "quote ")
            quote = strings.TrimPrefix(quote, "Quote ")
            Vote(s, m, quote)
        }

    }
}
