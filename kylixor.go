/* 	kylixor.go
_________________________________
Main code for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"bufio"
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
	Admin     string //Admin's Discord ID
	APIKey    string //Discord bot api key
	BootLogo  string //ASCII to display when starting up the bot
	Coins     string //Name of currency that bot uses (i.e. <gold> coins)
	Follow    bool   //Whether or not the bot joins/follows into voice channels for anthems
	LogID     string //ID of channel for logging
	Prefix    string //Prefix the bot will respond to
	ResetTime string //Time when dailies reset i.e. 19:00
	Status    string //Status of the bot (Playing <v1.0>)
}

var (
	config              = Config{}                 //Config structure from file
	currentVoiceChannel *discordgo.VoiceConnection //Current voice channel bot is in, nil if none
	self                *discordgo.User            //Discord user type of self (for storing bots user account)
	pwd, _              = os.Getwd()
)

//------------------------------------------------------------------------------
//-----------------               M A I N ( )               --------------------
//------------------------------------------------------------------------------
func main() {
	//Get random seed for later random number generation
	rand.Seed(time.Now().Unix())

	// Read in config file if exists
	if _, err := os.Stat(pwd + "/data/conf.json"); os.IsNotExist(err) {
		fmt.Println("\nCannot find conf.json, creating new...")
		InitConfFile()
		fmt.Println("\nPlease fill in the config.json file located in the data folder.")
		os.Exit(1)
	} else {
		//Check to make sure all needed config options are filled output and
		// Update config to account for any data structure changes
		config.UpdateConfig()

		//Check to see if bot token is provided
		if config.APIKey == "" {
			fmt.Println("No token provided. Please place your API key into the config.json file")
			return
		}

		//Check for defaults
		if config.ResetTime == "" {
			config.ResetTime = "20:00"
		}
	}

	// Read in user data file if exists
	if _, err := os.Stat(pwd + "/data/users.json"); os.IsNotExist(err) {
		fmt.Println("\nCannot find users.json, creating new...")
		InitUserFile()
	}

	// Reset all anthems
	jcc.ReadUserFile()
	for j := range jcc.Users {
		jcc.Users[j].PlayAnthem = true
	}
	jcc.WriteUserFile()

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

	// Register presenceUpdate to see who is online
	ky.AddHandler(presenceUpdate)

	// Register VoiceStateUpdate to check when users enter channel
	ky.AddHandler(voiceStateUpdate)

	// Open the websocket and begin listening for above events.
	err = ky.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	botChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(botChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	go func() {
		signalType := <-botChan
		fmt.Println(fmt.Sprintf("Shutting down nicely from signal type: %s", signalType))
		done <- true
	}()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Printf(config.BootLogo)
	fmt.Println("\nKylixor is now running.  Press CTRL-C to exit.")
	<-done

	// Cleanly close down the Discord session by disconnecting
	// from any connected voice channels
	fmt.Println("Disconnecting from voice channels...")
	if currentVoiceChannel != nil {
		if currentVoiceChannel.ChannelID != "" {
			currentVoiceChannel.Disconnect()
		}
	}
	fmt.Println("Closing websocket...")
	ky.Close()
	fmt.Println("Done...ending process...")
	os.Exit(0)
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	self = event.User

	//Set status once at start, then ticker takes over every hour
	SetStatus(s)

	// Start cronjobs
	go func() {
		gocron.Every(1).Day().At(config.ResetTime).Do(ResetDailies) //Reset dailies task
		<-gocron.Start()                                            //Start waiting for the cronjob
	}()

	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			SetStatus(s)
		}
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

	//Good Karma
	if strings.ToLower(m.Content) == "good bot" {
		jcc.Karma++
		jcc.WriteUserFile()
		s.MessageReactionAdd(m.ChannelID, m.ID, "😊")
	}

	//Bad Karma
	if strings.ToLower(m.Content) == "bad bot" {
		jcc.Karma--
		jcc.WriteUserFile()
		s.MessageReactionAdd(m.ChannelID, m.ID, "😞")
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

//----- I N I T I A L   S E T U P   F U N C T I O N S -----

//InitConfFile - Initialize config file if one is not found
func InitConfFile() {
	//Default values
	config.Prefix = "k!"
	config.ResetTime = "20:00"

	//Create and indent proper json output for the config
	configData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		panic(err)
	}

	//Create folder for data
	_ = os.Mkdir("data", 0755)

	// Open file
	jsonFile, err := os.Create(pwd + "/data/conf.json")
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
	err := decoder.Decode(&c)
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
	jsonFile, err := os.Create(pwd + "/data/conf.json")
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

//----- M I S C .   F U N C T I O N S -----

//ResetDailies - Function to call once a day to reset dailies
func ResetDailies() {
	for j := range jcc.Users {
		jcc.Users[j].Dailies = false
	}
	jcc.WriteUserFile()
}

//GetVersion - Get the version of the bot from the readme
func GetVersion(s *discordgo.Session) (ver string) {
	//Open the file and grab it line by line into textlines
	readme, err := os.Open("README.md")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(readme)
	scanner.Split(bufio.ScanLines)
	var textlines []string

	for scanner.Scan() {
		textlines = append(textlines, scanner.Text())
	}

	//Second line of the readme will always be the version number
	ver = textlines[1]

	//Close file and return version number
	readme.Close()
	return ver
}

//SetStatus - sets the status of the bot to the version and the default help commands
func SetStatus(s *discordgo.Session) {
	s.UpdateStatus(0, fmt.Sprintf("%s - %shelp", GetVersion(s), config.Prefix))
}
