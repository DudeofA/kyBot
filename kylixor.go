package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"sort"
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
	Admin       string
	Bots        []string
	Coins       string
	DefaultChan string
	LogID       string
	LogMessage  bool
	LogStatus   bool
	LogVoice    bool
	Monitor     []string
	Noise       bool
	Prefix      string
	Status      string
	Test        []string
}

var config = Config{}
var self *discordgo.User
var token string
var curChan *discordgo.VoiceConnection

//Read in the config into the Config structure
func (c Config) ReadConfig() {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&c)
	if err != nil {
		fmt.Println("error: ", err)
	}
	file.Close()
}

//Write out the current Config structure to file, indented nicely
func (c Config) WriteConfig() {
	//Indent so its readable
	configData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		panic(err)
	}
	//Open file
	jsonFile, err := os.Create("conf.json")
	if err != nil {
		panic(err)
	}
	//Write to file
	_, err = jsonFile.Write(configData)
	if err != nil {
		panic(err)
	}
	//Cleanup
	jsonFile.Close()
}

//Function to call once a day
func ResetDailies() {
	for j := range USArray.Users {
		USArray.Users[j].Dailies = false
	}
	USArray.WriteUserFile()
}

func main() {
	//Read in files
	rand.Seed(time.Now().Unix())

	if _, err := os.Stat("users.json"); os.IsNotExist(err) {
		InitUserFile()
	}
	config.ReadConfig()
	config.WriteConfig()
	USArray.ReadUserFile()

	for j := range USArray.Users {
		USArray.Users[j].PlayAnthem = true
	}
	USArray.WriteUserFile()

	go func() {
		gocron.Every(1).Day().At("20:00").Do(ResetDailies)
		<-gocron.Start()
	}()

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

	// Register ready as a callback for the ready events
	ky.AddHandlerOnce(ready)

	// Register messageCreate as a callback for the messageCreate events
	ky.AddHandler(messageCreate)

	// Register messageDelete as a callback for the messageDelete events
	ky.AddHandler(messageDelete)

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
	if curChan != nil {
		if curChan.ChannelID != "" {
			curChan.Disconnect()
		}
	}
	ky.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	s.UpdateStatus(0, config.Status)
	self = event.User
	USArray.GID = event.Guilds[0].ID
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

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// This function will be called when a user changes their voice state
// (mute, deafen, join channel, leave channel, etc.)
func VoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	//Update the user and if it is a real update log it
	VoiceChannelChange := USArray.UpdateUser(s, v, "VOICE")
	if VoiceChannelChange {
		Log(s, v, "VOICE")
		//If the VoiceStateUpdate is a join channel
		g, _ := s.Guild(USArray.GID)
		if v.ChannelID != g.AfkChannelID {
			if v.ChannelID != "" {
				usr, _ := USArray.ReadUser(s, v, "VOICE")
				if usr.PlayAnthem && usr.Anthem != "" {
					PlayAnthem(s, v, usr.Anthem)
					time.Sleep(3 * time.Second)
				}
			}
		}
		//Join channel with most people
		if len(g.VoiceStates) > 0 {
			m := make(map[string]int)
			for i := range g.VoiceStates {
				m[g.VoiceStates[i].ChannelID] += 1
			}

			pl := make(PairList, len(m))
			i := 0
			for k, v := range m {
				pl[i] = Pair{k, v}
				i++
			}
			sort.Sort(sort.Reverse(pl))

			//If bot is the only one left, leave
			if pl[0].Value == 1 && curChan != nil {
				curChan.Disconnect()
				curChan.ChannelID = ""
			} else {
				//Join channel with most people
				curChan, _ = s.ChannelVoiceJoin(USArray.GID, pl[0].Key, false, false)
			}
		}
	}
}

func messageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	if m.ChannelID != config.LogID {
		Log(s, m, "DEL")
	}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bots
	// This isn't required in this specific example but it's a good practice.
	if m.Author.Bot {
		return
	}

	// Log every message into the log channel
	if m.ChannelID != config.LogID {
		Log(s, m, "MSG")
	}

	if m.Content == "(╯°□°）╯︵ ┻━┻" {
		s.ChannelMessageSend(m.ChannelID, "┬─┬ノ( º _ ºノ)")
	}

	if strings.HasPrefix(m.Content, config.Prefix) {
		// Remove prefix
		input := strings.TrimPrefix(m.Content, config.Prefix)
		// Split message into command and anything after
		command := strings.SplitN(input, " ", 2)

		switch strings.ToLower(command[0]) {

		case "account":
			usr, _ := USArray.ReadUser(s, m, "MSG")
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("💵 | You have a total of **%d** %s coins", usr.NoiseCredits, config.Coins))
			break

		case "anthem":
			SetAnthem(s, m)
			break

		case "bitconnect":
			PlayClip(s, m, "bitconnect")
			break

		case "dailies":
			_, i := USArray.ReadUser(s, m, "MSG")
			usr := &USArray.Users[i]
			if !usr.Dailies {
				usr.NoiseCredits += 100
				usr.Dailies = true
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
					"💵 | Dailies received! Total %s coins: **%d**",
					config.Coins, usr.NoiseCredits))
			} else {
				_, nextRuntime := gocron.NextRun()
				timeUntil := time.Until(nextRuntime)
				hour := timeUntil / time.Hour
				timeUntil -= hour * time.Hour
				min := timeUntil / time.Minute
				timeUntil -= min * time.Minute
				sec := timeUntil / time.Second
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
					"💵 | You have already collected today's dailies.\nDailies reset in %d hour(s), %d minute(s) and %d second(s).",
					hour, min, sec))
			}
			USArray.WriteUserFile()
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

		case "quote":
			Vote(s, m, command[1])
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
				config.ReadConfig()
				USArray.ReadUserFile()
				s.ChannelMessageSend(m.ChannelID, "Config reloaded")
			}
			break

		case "remindme":
			_, i := USArray.ReadUser(s, m, "MSG")

			//If there are no reminders
			if USArray.Users[i].Reminders == nil {
				USArray.Users[i].Reminders = make([]string, 0)
			}
			USArray.Users[i].Reminders = append(USArray.Users[i].Reminders, m.Content)
			USArray.WriteUserFile()
			break

		case "status":
			if m.Author.ID == config.Admin {
				config.ReadConfig()
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
			responseList := make([]string, 0)
			responseList = append(responseList,
				"My slave isn't that good at coding, sorry",
				"I'm trying really hard but you're not being very clear",
				"I have no idea what you're talking about",
				"You think I'm some kind of AI?",
				"I'm smarter than you but I'm not _that_ smart",
				"I don't want to do that",
				"No")

			s.ChannelMessageSend(m.ChannelID, responseList[rand.Intn(len(responseList))])
		}
	}
}
