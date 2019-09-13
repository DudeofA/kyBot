/* 	kylixor.go
_________________________________
Main code for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"
)

// ----- GLOBAL VARIABLES -----

//BotConfig - Global bot config
type BotConfig struct {
	Admin     string `json:"admin"`     // Admin's Discord ID
	APIKey    string `json:"apiKey"`    // Discord bot api key
	BootLogo  string `json:"bootLogo"`  // ASCII to display when starting up the bot
	DailyAmt  int    `json:"dailyAmt"`  // Amount of dailies to be collected daily
	DBConfig  DBConf `json:"dbConfig"`  // Database configuration
	LogID     string `json:"logID"`     // ID of channel for logging
	Prefix    string `json:"prefix"`    // Prefix the bot will respond to
	ResetTime string `json:"resetTime"` // Time when dailies reset i.e. 19:00
	Status    string `json:"status"`    // Status of the bot (Playing <v1.0>)
	Version   string `json:"version"`   // Current version of the bot
}

// DBConf - Holds all data needed to connect and use a mongoDB database
type DBConf struct {
	DBName string `json:"dbName"` // Name of database where data is kept
	URI    string `json:"dbURI"`  // URI of database to connect to (localhost or hosted)
}

var (
	botConfig           = BotConfig{}              // Global bot config
	currentVoiceChannel *discordgo.VoiceConnection // Current voice channel bot is in, nil if none
	self                *discordgo.User            // Discord user type of self (for storing bot's user account)
	pwd                 string                     // Current working directory
)

//------------------------------------------------------------------------------
//-----------------               M A I N ( )               --------------------
//------------------------------------------------------------------------------
func main() {
	fmt.Println("Starting kylixor bot...")

	// Set pwd to the directory of the bot's files
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	bin := filepath.Dir(ex)
	pwd = filepath.Dir(bin)

	// Get random seed for later random number generation
	rand.Seed(time.Now().UTC().UnixNano())

	// Read in config file if exists
	if _, err = os.Stat(filepath.FromSlash(pwd + "/conf.json")); os.IsNotExist(err) {
		fmt.Println("\nCannot find conf.json, creating new...")
		InitBotConfFile()
		fmt.Println("\nPlease fill in the config.json file located in the data folder.")
		os.Exit(1)
	} else {
		botConfig.Read()

		// Check DB URI
		if botConfig.DBConfig.URI == "" {
			fmt.Println("Database URI not provided.  Please place your database URI into the config.json file")
			botConfig.Write()
			return
		}

		// Check to see if bot token is provided
		if botConfig.APIKey == "" {
			fmt.Println("No token provided. Please place your API key into the config.json file")
			botConfig.Write()
			return
		}

		if botConfig.DBConfig.DBName == "" {
			fmt.Println("Please provide the correct database name for this instance in the config.json file")
			botConfig.Write()
			return
		}

		// Default mandatory values

		if botConfig.ResetTime == "" {
			botConfig.ResetTime = "20:00"
		}
		if botConfig.DailyAmt == 0 {
			botConfig.DailyAmt = 100
		}
		if botConfig.Prefix == "" {
			botConfig.Prefix = "k!"
		}

		// Save version
		botConfig.Version = GetVersion()

		botConfig.Write()
	}

	// Connect/Setup database
	kdb.Init()

	// Check for dictionary file
	if runtime.GOOS == "windows" {
		_, err = os.Stat(filepath.FromSlash(pwd + "/dict/words.txt"))
		if err != nil {
			fmt.Println("No dictionary file supplied, please add words.txt to the dict folder, full of words to use for hangman")
			return
		}
	} else {
		_, err = os.Stat(filepath.FromSlash("/usr/share/dict/words"))
		if err != nil {
			fmt.Println("Error accessing dictionary at /usr/share/dict/words")
			return
		}
	}

	// Create a new Discord session using the provided bot token.
	ky, err := discordgo.New("Bot " + botConfig.APIKey)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	//Use the state to cache messages
	ky.State.MaxMessageCount = 100

	// Register ready for the ready event
	ky.AddHandlerOnce(Ready)

	// Register messageCreate for the messageCreate events
	ky.AddHandler(MessageCreate)

	// Register presenceUpdate to see who is online
	ky.AddHandler(PresenceUpdate)

	// Register VoiceStateUpdate to check when users enter channel
	ky.AddHandler(VoiceStateUpdate)

	// Register MessageAddReaction to check for reactions (for hangman)
	ky.AddHandler(MessageReactionAdd)

	//BootLogo
	fmt.Println(botConfig.BootLogo)

	// Open the websocket and begin listening for above events.
	err = ky.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}
	fmt.Println("Successfully opened Discord bot session!")

	//Create channels to watch for kill signals
	botChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	//Bot will end on any of the following signals
	signal.Notify(botChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	go func() {
		signalType := <-botChan
		fmt.Println(fmt.Sprintf("\nShutting down from signal: %s", signalType))
		done <- true
	}()

	// Wait here until CTRL-C or other term signal is received.
	<-done

	// Cleanly close down the Discord session by disconnecting
	// from any connected voice channels
	fmt.Println("Disconnecting from voice channels...")
	if currentVoiceChannel != nil {
		if currentVoiceChannel.ChannelID != "" {
			currentVoiceChannel.Disconnect()
		}
	}
	fmt.Println("Closing discord websocket...")
	ky.Close()
	fmt.Println("Closing database connection...")
	mdbClient.Disconnect(context.TODO())
	fmt.Println("Done...ending process...goodbye...")
	os.Exit(0)
}

// Ready - This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func Ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	self = event.User

	servers := s.State.Guilds
	fmt.Printf("Kylixor discord bot has started on %d servers\n", len(servers))

	// Start cronjobs
	go func() {
		gocron.Every(1).Day().At(botConfig.ResetTime).Do(ResetDailies) //Reset dailies task
		<-gocron.Start()                                               //Start waiting for the cronjob
	}()

	// Set status once at start, then ticker takes over every hour
	SetStatus(s)

	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			SetStatus(s)
		}
	}()

	LogTxt(s, "INFO", "Bot starting up...")

	fmt.Println("\nKylixor discord bot is now running.  Press CTRL-C to exit.")
}

// MessageReactionAdd - Called whenever a message is sent to the discord
func MessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == self.ID {
		return
	}

	//If reaction is on a message that is a hangman game, guess that reaction
	// for i := range kdb.Servers {
	// 	if kdb.Servers[i].HM.Message == r.MessageID {
	// 		ReactionGuess(s, r, &kdb.Servers[i].HM)
	// 	}
	// }
}

// PresenceUpdate - Called when any user changes their status (online, away, playing a game, etc)
func PresenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {

}

// VoiceStateUpdate - Called whenever a user changes their voice state (muted, deafen, connected, disconnected)
func VoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	LogVoice(s, v)
}

// MessageCreate - Called whenever a message is sent to the discord
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Get Guild index to use later on
	guild := kdb.ReadGuild(s, m.GuildID)

	// Return if the message was sent by a bot to avoid infinite loops
	if m.Author.Bot || m.ChannelID == botConfig.LogID {
		return
	}

	// Log message to log channel for debugging
	LogMsg(s, m)

	// Fix any flipped tables
	if m.Content == "(╯°□°）╯︵ ┻━┻" {
		s.ChannelMessageSend(m.ChannelID, "┬─┬ノ( º _ ºノ)")
	}

	// Good Karma
	if strings.ToLower(m.Content) == "good bot" {
		guild.Karma++
		guild.Update(s)
		s.MessageReactionAdd(m.ChannelID, m.ID, "😊")
	}

	// Bad Karma
	if strings.ToLower(m.Content) == "bad bot" {
		guild.Karma--
		guild.Update(s)
		s.MessageReactionAdd(m.ChannelID, m.ID, "😞")
	}

	if m.Content == fmt.Sprintf("<@%s>", self.ID) {
		err := s.MessageReactionAdd(m.ChannelID, m.ID, "👋")
		if err != nil {
			panic(err)
		}
	}

	// If the message sent is a command with the set prefix
	if strings.HasPrefix(m.Content, guild.Config.Prefix) {
		// Trim the prefix to extract the command
		input := strings.TrimPrefix(m.Content, guild.Config.Prefix)
		// Split command into the command and what comes after
		inputPieces := strings.SplitN(input, " ", 2)
		command := strings.ToLower(inputPieces[0])
		var data string
		if len(inputPieces) == 2 {
			data = inputPieces[1]
		}

		// Send the data to the command function for execution
		runCommand(s, m, command, data)
	}
}

//----- I N I T I A L   S E T U P   F U N C T I O N S -----

// InitBotConfFile - Initialize config file if one is not found
func InitBotConfFile() {
	// Create and indent proper json output for the config
	configData, err := json.MarshalIndent(botConfig, "", "    ")
	if err != nil {
		panic(err)
	}

	// Create folder for data
	_ = os.Mkdir("data", 0755)

	// Open file
	jsonFile, err := os.Create(filepath.FromSlash(pwd + "/data/conf.json"))
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

// ReadBotConfig - Read in config file into Config structure
func (c *BotConfig) Read() {
	file, _ := os.Open(filepath.FromSlash(pwd + "/data/conf.json"))
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&c)
	if err != nil {
		panic(err)
	}
	file.Close()
}

// WriteBotConfig - Write out the current Config structure to file, indented nicely
func (c *BotConfig) Write() {
	// Indent so its readable
	configData, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		panic(err)
	}
	// Open file
	jsonFile, err := os.Create(filepath.FromSlash(pwd + "/data/conf.json"))
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

// Update - update configuration file by reading then writing
// Updates config file to correct syntax
func (c *BotConfig) Update() {
	botConfig.Read()
	botConfig.Write()
}

//----- M I S C .   F U N C T I O N S -----

// ResetDailies - Function to call once a day to reset dailies
func ResetDailies() {
	// for i := range kdb.Users {
	// 	kdb.Users[i].DoneDailies = false
	// }
}

// GetVersion - Get the version of the bot from the readme
func GetVersion() (ver string) {
	// Open the file and grab it line by line into textlines
	readme, err := os.Open(filepath.FromSlash(pwd + "/README.md"))
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(readme)
	scanner.Split(bufio.ScanLines)
	var textlines []string

	for scanner.Scan() {
		textlines = append(textlines, scanner.Text())
	}

	// Second line of the readme will always be the version number
	if len(textlines) < 2 {
		panic("Version needs to be in the second line of the README in format: <v#.#.#")
	} else {
		ver = textlines[1]
	}

	// Close file, save version number, and return version number
	readme.Close()
	botConfig.Version = ver
	return ver
}

// SetStatus - sets the status of the bot to the version and the default help commands
func SetStatus(s *discordgo.Session) {
	s.UpdateStatus(0, fmt.Sprintf("%shelp - %s", botConfig.Prefix, botConfig.Version))
}

// CheckAdmin - returns true if user is admin, otherwise posts that permission is denied
func CheckAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if botConfig.Admin == m.Author.ID {
		return true
	}

	s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command")
	return false
}

// MemberHasPermission - Checks if the user has permission to do the given action in the given channels
func MemberHasPermission(s *discordgo.Session, guildID string, userID string, permission int) (bool, error) {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.GuildMember(guildID, userID); err != nil {
			return false, err
		}
	}

	// Iterate through the role IDs stored in member.Roles to check permissions
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			return false, err
		}
		if role.Permissions&permission != 0 {
			return true, nil
		}
	}

	return false, nil
}

// CreationTime returns the creation time of a Snowflake ID relative to the creation of Discord.
// Taken from https://github.com/Moonlington/FloSelfbot/blob/master/commands/commandutils.go#L117
func CreationTime(ID string) (t time.Time, err error) {
	i, err := strconv.ParseInt(ID, 10, 64)
	if err != nil {
		return
	}
	timestamp := (i >> 22) + 1420070400000
	t = time.Unix(timestamp/1000, 0)
	return
}

// GetAge - Take in a string that should be a discord snowflake ID, then calulate, format, and return the age
func GetAge(rawID string) string {

	id := rawID
	id = strings.TrimPrefix(id, "<#")
	id = strings.TrimPrefix(id, "<@")
	id = strings.TrimPrefix(id, "!")
	id = strings.TrimSuffix(id, ">")
	// Attempt to get the creation time of the ID given
	t, err := CreationTime(id)
	if err != nil {
		return fmt.Sprintf("Not a valid Discord Snowflake ID: \"%s\"", rawID)
	}

	day := time.Hour * 24
	year := 365 * day
	var years time.Duration

	tAlive := time.Now().Sub(t)

	if tAlive >= year {
		// At least a year
		years = tAlive / year
		tAlive -= years * year
	}

	days := tAlive / day
	tAlive -= days * day

	// If the age is more than a year
	tYears := fmt.Sprintf("%d years, ", years)
	and := "and "
	tDays := fmt.Sprintf("%d days, ", days)

	return fmt.Sprintf("%s was created on %s. They are %s%s%s%d hours old.",
		rawID, t.Format("Jan 2 3:04:05PM 2006"), tYears, tDays, and, int(tAlive.Hours()))
}

// QuotePrint - Prints the quote with nice colors
func QuotePrint(s *discordgo.Session, m *discordgo.MessageCreate, q Quote) *discordgo.Message {
	// Format quote with colors using the CSS formatting
	fmtQuote := fmt.Sprintf("```ini\n[ %s ] - [ %s ]\n%s\n```", q.Timestamp.Format("Jan 2 3:04:05PM 2006"), q.Identifier, q.Quote)
	// Return the sent message for vote monitoring
	msg, err := s.ChannelMessageSend(m.ChannelID, fmtQuote)
	if err != nil {
		panic(err)
	}
	return msg
}
