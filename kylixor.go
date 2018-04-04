package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/jasonlvhit/gocron"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

type Config struct {
	Admin      string
	Bots       []string
	LogID      string
	LogMessage bool
	LogStatus  bool
	LogVoice   bool
	Monitor    []string
	Noise      bool
	Prefix     string
	Status     string
	Test       []string
}

var config = Config{}
var followUser *discordgo.User
var token string
var stopChan = make(chan bool)

func ReadConfig() {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error: ", err)
	}
	file.Close()
}

func ResetDailies() {
	for j := range USArray.Users {
		USArray.Users[j].Dailies = false
	}
	WriteUserFile()
}

func main() {
	//Read in files
	if _, err := os.Stat("users.json"); os.IsNotExist(err) {
		InitUserFile()
	}
	ReadConfig()
	ReadUserFile()

	gocron.Every(1).Day().At("00:00").Do(ResetDailies)
	<-gocron.Start()

	//Account for no token at runtime
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

	// Register presenceUpdate to see who is online
	ky.AddHandler(presenceUpdate)

	// Register VoiceStateUpdate to check when users enter channel
	ky.AddHandler(VoiceStateUpdate)

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
	s.UpdateStatus(0, config.Status)
}

// This function will be called each time certain (or all) users change their
// online status
func presenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	//Go through the range of who to monitor and log if needed
	for _, b := range config.Monitor {
		if b == p.User.ID {
			Log(s, p, "STATUS")
		}
	}
}

// This function will be called when a user changes their voice state
// (mute, deafen, join channel, leave channel, etc.)
func VoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	//Update the user and if it is a real update log it
	success := UpdateUser(s, v, "VOICE")
	if success {
		Log(s, v, "VOICE")
	}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bots
	// This isn't required in this specific example but it's a good practice.
	for _, b := range config.Bots {
		if b == m.Author.ID {
			return
		}
	}

	// Log every message into the log channel
	if m.ChannelID != config.LogID {
		Log(s, m, "MSG")
	}

	if strings.HasPrefix(m.Content, config.Prefix) {
		// Remove prefix for 'performance'
		input := strings.TrimPrefix(m.Content, config.Prefix)

		switch input {

		case "bitconnect":
			PlayClip(s, m, "bitconnect")
			break

		case "dailies":
			_, i := ReadUser(s, m, "MSG")
			usr := &USArray.Users[i]
			if !usr.Dailies {
				usr.NoiseCredits += 1
				usr.Dailies = true
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Dailies received! Total credits: %d", usr.NoiseCredits))
			} else {
				s.ChannelMessageSend(m.ChannelID, "Already got dailies, check back tomorrow.")
			}
			WriteUserFile()
			break

		case "help":
			readme, err := ioutil.ReadFile("README.md")
			if err != nil {
				panic(err)
			}
			// Print readme in code brackets so it doesn't look awful
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```"+string(readme)+"```"))
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
			// List all quote, with lag
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
				ReadUserFile()
				s.ChannelMessageSend(m.ChannelID, "Config reloaded")
			}
			break

		case "status":
			if m.Author.ID == config.Admin {
				ReadConfig()
				s.UpdateStatus(0, config.Status)
				s.ChannelMessageSend(m.ChannelID, "Status refreshed")
			}
			break

		case "test":
			if m.Author.ID == config.Admin {
				Test(s, m)
			}
			break

		case "yee":
			PlayClip(s, m, "yee")
			break

		default:
			s.ChannelMessageSend(m.ChannelID, "Not a command I'm pretty sure")
		}

		if strings.HasPrefix(m.Content, "quote ") {
			//NEEDS IMPROVEMENTS
			quote := strings.TrimPrefix(m.Content, "quote ")
			Vote(s, m, quote)
		}

	}
}
