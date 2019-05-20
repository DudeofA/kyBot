/* 	kylixor.go
_________________________________
Main code for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"
)

// ----- GLOBAL VARIABLES -----

//Config - structure to hold variables from the config file
type Config struct {
	Admin  string //Admin's Discord ID
	APIKey string //Discord bot api key
	Coins  string //Name of currency that bot uses (i.e. <gold> coins)
	Follow bool   //Whether or not the bot joins/follows into voice channels for anthems
	LogID  string //ID of channel for logging
	Prefix string //Prefix the bot will respond to
	Status string //Status of the bot (Playing <v1.0>)
}

var (
	config              = Config{}                 //Config structure from file
	currentVoiceChannel *discordgo.VoiceConnection //Current voice channel bot is in, nil if none
	self                *discordgo.User            //Discord user type of self (for storing bots user account)
	err                 error                      //One error to rule them all
	pwd, _              = os.Getwd()
)

//InitConfFile - Initialize config file if one is not found
func InitConfFile() {
	//Default values
	config.Prefix = "k!"
	config.Status = "k!help"

	//Create and indent proper json output for the config
	configData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		panic(err)
	}

	//Create folder for data
	_ = os.Mkdir("data", 0755)

	// Open file
	jsonFile, err := os.Create(pwd + "/data/config.json")
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

//ReadConfig - Read in config file into Config structure
func (c *Config) ReadConfig() {
	file, _ := os.Open(pwd + "/data/conf.json")
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		panic(err)
	}
	file.Close()
}

//WriteConfig - Write out the current Config structure to file, indented nicely
func (c *Config) WriteConfig() {
	//Indent so its readable
	configData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		panic(err)
	}
	//Open file
	jsonFile, err := os.Open(pwd + "/data/conf.json")
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

//UpdateConfig - update configuration file by reading then writing
//Updates config file to correct syntax
func (c *Config) UpdateConfig() {
	config.ReadConfig()
	config.WriteConfig()
}

//ResetDailies - Function to call once a day to reset dailies
func ResetDailies() {
	// for j := range USArray.Users {
	// 	USArray.Users[j].Dailies = false
	// }
	// USArray.WriteUserFile()
}

//------------------------------------------------------------------------------
//-----------------               M A I N ( )				--------------------
//------------------------------------------------------------------------------
func main() {
	//Get random seed for later random number generation
	rand.Seed(time.Now().Unix())

	// Read in config file if exists
	if _, err := os.Stat(pwd + "/data/conf.json"); os.IsNotExist(err) {
		fmt.Println("\nCannot find conf.json, creating new...")
		InitConfFile()
	}

	// Update config to account for any data structure changes
	config.UpdateConfig()

	// Read in user data file if exists
	// if _, err := os.Stat("data/users.json"); os.IsNotExist(err) {
	// fmt.Println("\nCannot find users.json, creating new...")
	// InitUserFile()
	// }

	// Reset all anthems
	// USArray.ReadUserFile()
	// for j := range USArray.Users {
	// 	USArray.Users[j].PlayAnthem = true
	// }
	// USArray.WriteUserFile

	//Check to see if bot token is provided
	if config.APIKey == "" {
		fmt.Println("No token provided. Please place your API key into the config.json file")
		return
	}

	// Create a new Discord session using the provided bot token.
	ky, err := discordgo.New("Bot " + config.APIKey)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready for the ready event
	ky.AddHandlerOnce(ready)

	// Register messageCreate for the messageCreate events
	ky.AddHandler(messageCreate)

	// Register messageDelete for the messageDelete events
	// ky.AddHandler(messageDelete)

	// Register presenceUpdate to see who is online
	ky.AddHandler(presenceUpdate)

	// Register VoiceStateUpdate to check when users enter channel
	ky.AddHandler(voiceStateUpdate)

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
	if currentVoiceChannel != nil {
		if currentVoiceChannel.ChannelID != "" {
			currentVoiceChannel.Disconnect()
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
	// USArray.GID = event.Guilds[0].ID

	// Start cronjobs
	go func() {
		gocron.Every(1).Day().At("19:00").Do(ResetDailies) //Reset dailies task

		<-gocron.Start() //Start waiting for the cronjobs
	}()
}

//presenceUpdate - Called when any user changes their status (online, away, playing a game, etc)
func presenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {

}

//voiceStateUpdate - Called whenever a user changes their voice state (muted, deafen, connected, disconnected)
func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {

}

//messageCreate - Called whenever a message is sent to the discord
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	//Return if the message was sent by a bot to avoid infinite loops
	if m.Author.Bot {
		return
	}

	//Fix any flipped tables
	if m.Content == "(╯°□°）╯︵ ┻━┻" {
		s.ChannelMessageSend(m.ChannelID, "┬─┬ノ( º _ ºノ)")
	}

	//If the message sent is a command with the set prefix
	if strings.HasPrefix(m.Content, config.Prefix) {
		//Trim the prefix to extract the command
		input := strings.TrimPrefix(m.Content, config.Prefix)
		//Split command into the command and what comes after
		inputPieces := strings.SplitN(input, " ", 2)
		command := strings.ToLower(inputPieces[0])
		var data string
		if len(inputPieces) == 2 {
			data = inputPieces[1]
		}

		//Send the data to the command function for execution
		runCommand(s, m, command, data)

	}

}
