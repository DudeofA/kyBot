/* 	kylixor.go
	_________________________________
	Main code for Kylixor Discord Bot
	Andrew Langhill
	kylixor.com
*/

package main

import(
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

type Config struct {
	Admin	string
	Coins	string 		//Name of currency that bot uses (i.e. <gold> coins)
    Follow      bool 	//Whether or not the bot joins/follows into voice channels for anthems
	LogID       string 	//ID of channel for logging
	Noise       bool 	//Whether the bot will use function that play sound
	Prefix      string 	//Prefix the bot will respond to
	Status      string	//Status of the bot (Playing <v1.0>)
}

// ----- GLOBAL VARIABLES -----
var currentVoiceChannel *discordgo.VoiceConnection 	//Current voice channel bot is in, nil if none
var config = Config{}								//Config structure from file
var self *discordgo.User 							//discord user type of self (bots user account)
var APItoken string 								//API token from flag

func InitConfFile() {
	config.Prefix = "k!"
	config.Status = "k!help"

	configData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		panic(err)
	}
	// Open file
	jsonFile, err := os.Create("data/conf.json")
	if err != nil {
		panic(err)
	}
	// Write to file
	_, err = jsonFile.Write(configData)
	if err != nil {
		panic(err)
	}
	// Cleanup
	jsonFile.Close()
}

// Read in the config into the Config structure
func (c *Config) ReadConfig() {
	file, _ := os.Open("data/conf.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&c)
	if err != nil {
		panic(err)
	}
	file.Close()
}

// Write out the current Config structure to file, indented nicely
func (c *Config) WriteConfig() {
	//Indent so its readable
	configData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		panic(err)
	}
	//Open file
	jsonFile, err := os.Create("data/conf.json")
	if err != nil {
		panic(err)
	}
	// Write to file
	_, err = jsonFile.Write(configData)
	if err != nil {
		panic(err)
	}
	// Cleanup
	jsonFile.Close()
}

func (c *Config) UpdateConfig() {
	config.ReadConfig();
	config.WriteConfig();
}

// Function to call once a day
func ResetDailies() {
	for j := range USArray.Users {
		USArray.Users[j].Dailies = false
	}
	USArray.WriteUserFile()
}

func main() {

	rand.Seed(time.Now().Unix())

	// Parse bot token
	flag.StringVar(&APItoken, "t", "", "Bot Token")
	flag.Parse()
	if token == "" {
		fmt.Println("No token provided. Please run: kylixor -t <bot token>")
		return
	}

	// Read in config file if exists
	if _, err := os.Stat("data/conf.json"); os.IsNotExist(err) {
		fmt.Println("\nCannot find conf.json, creating new...")
		InitConfFile()
	}

	// Update config to account for any data structure changes
	config.UpdateConfig();

	// Read in user data file if exists
	if _, err := os.Stat("data/users.json"); os.IsNotExist(err) {
		fmt.Println("\nCannot find users.json, creating new...")
		InitUserFile()
	}

    // Reset all anthems
	USArray.ReadUserFile()
	for j := range USArray.Users {
		USArray.Users[j].PlayAnthem = true
	}
	USArray.WriteUserFile()

	// Start cronjob to reset dailies every day at 7pm
	go func() {
		gocron.Every(1).Day().At("19:00").Do(ResetDailies)

		<-gocron.Start()
	}()

	// Create a new Discord session using the provided bot token.
	ky, err := discordgo.New("Bot " + APItoken)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready for the ready event
	ky.AddHandlerOnce(ready)

	// Register messageCreate for the messageCreate events
	ky.AddHandler(messageCreate)

	// Register messageDelete for the messageDelete events
	ky.AddHandler(messageDelete)

	// Register presenceUpdate to see who is online
	ky.AddHandler(presenceUpdate)

	// Register VoiceStateUpdate to check when users enter channel
	ky.AddHandler(VoiceStateUpdate)

	// Open the websocket and begin listening for above events.
	err = ky.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Kylixor is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session by disconnecting
	// from any connected voice channels
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
