/* 	kylixor.go
_________________________________
Main code for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
)

// ----- GLOBAL VARIABLES -----

//K - Kylixor bot
type K struct {
	botConfig BotConfig          // Bot's global configuration
	cron      *cron.Cron         // Cronjob for dailies timer
	db        *sql.DB            // Raw database connection bundle
	kdb       KDB                // Server/User database
	logfile   *os.File           // Log file handle
	session   *discordgo.Session // Session info
	state     BotState           // Volitile state of bot
}

//BotConfig - Global bot config
type BotConfig struct {
	Admin     string `json:"admin"`     // Admin's Discord ID
	APIKey    string `json:"apiKey"`    // Discord bot api key
	BootLogo  string `json:"bootLogo"`  // ASCII to display when starting up the bot
	DailyAmt  int    `json:"dailyAmt"`  // Amount of dailies to be collected daily
	DBName    string `json:"dbName"`    // Name of database where data is kept
	DBURI     string `json:"dbURI"`     // URI of database to connect to (localhost or hosted)
	MinVotes  int    `json:"minVotes"`  // Minimum votes to pass a vote
	Prefix    string `json:"prefix"`    // Prefix the bot will respond to
	ResetTime string `json:"resetTime"` // Time when dailies reset i.e. 19:00
	Status    string `json:"status"`    // Status of the bot (Playing <v1.0>)
	Version   string `json:"version"`   // Current version of the bot
}

//BotState - Current state of the bot (polled at startup)
type BotState struct {
	currentVC discordgo.VoiceConnection // Current voice channel connected to
	pwd       string                    // Working directory
	servers   int                       // Servers listenning on
	self      *discordgo.User           // Self object
	version   string                    // Identify own version
}

//------------------------------------------------------------------------------
//-----------------               M A I N ( )               --------------------
//------------------------------------------------------------------------------
var k K // Bot

func main() {
	var err error

	// Set pwd to the directory of the bot's files
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	bin := filepath.Dir(ex)
	k.state.pwd = filepath.Dir(bin)

	// Create logfile if needed, then get its handle
	k.logfile, err = os.OpenFile(filepath.FromSlash(k.state.pwd+"/k.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer k.logfile.Close()

	k.logfile.WriteString("\n\n\n")
	k.Log("STARTUP", "Starting kylixor bot...")

	// Get random seed for later random number generation
	rand.Seed(time.Now().UTC().UnixNano())

	// Read in config file if exists
	if _, err = os.Stat(filepath.FromSlash(k.state.pwd + "/conf.json")); os.IsNotExist(err) {
		fmt.Println("Cannot find conf.json, creating new...")
		k.botConfig.Init()
		fmt.Println("\nPlease fill in the config.json file located in the data folder.")
		os.Exit(1)
	} else {
		k.botConfig.Read()

		// Default mandatory values
		if k.botConfig.ResetTime == "" {
			k.botConfig.ResetTime = "20:00"
		}
		if k.botConfig.DailyAmt == 0 {
			k.botConfig.DailyAmt = 100
		}
		if k.botConfig.MinVotes == 0 {
			k.botConfig.MinVotes = 3
		}
		if k.botConfig.Prefix == "" {
			k.botConfig.Prefix = "k!"
		}

		// Check DB URI
		if k.botConfig.DBURI == "" {
			fmt.Println("Database URI not provided.  Please place your database URI into the config.json file")
			k.botConfig.Write()
			return
		}

		// Check to see if bot token is provided
		if k.botConfig.APIKey == "" {
			fmt.Println("No token provided. Please place your API key into the config.json file")
			k.botConfig.Write()
			return
		}

		// Check for database name
		if k.botConfig.DBName == "" {
			fmt.Println("Please provide the correct database name for this instance in the config.json file")
			k.botConfig.Write()
			return
		}

		// Write back config
		k.botConfig.Write()
	}

	// Save version
	k.state.version = GetBotVersion()

	// Connect/Setup database
	k.Log("KDB", "Connecting to MySQL DB...")

	// Connect to MySQL server
	k.db, err = sql.Open("mysql", k.botConfig.DBURI+"/"+k.botConfig.DBName)
	if err != nil {
		k.Log("FATAL", "Failed to get database connection set up")
		panic(err)
	}
	defer k.db.Close()

	// Test connection to server
	err = k.db.Ping()
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			fmt.Println("Cannot connect to SQL server, is SQL running?")
			panic(err)
		}
		panic(err)
	}

	// Create database if not found
	var row string
	result := k.db.QueryRow("SHOW TABLES LIKE 'state'")
	err = result.Scan(&row)
	switch err {
	case sql.ErrNoRows:
		k.Log("WARN", "No state table found, creating new...")
		k.kdb.Init()
	case nil:
		var version string
		result := k.db.QueryRow("SELECT version FROM state")
		err = result.Scan(&version)
		if err != nil {
			k.Log("WARN", "Error reading database version")
		}
		k.Log("INFO", "Found database version: "+version)
		localVer := k.state.version
		if version != localVer {
			k.Log("WARN", "Database version: (\""+version+"\") does not equal bot's version: (\""+localVer+"\"), updating KDB if necessary...")
			k.kdb.Update(version)
		}
		break
	default:
		panic(err)
	}

	k.Log("KDB", fmt.Sprintf("Connected to MySQL - %s", k.botConfig.DBName))

	// Check for dictionary file
	if runtime.GOOS == "windows" {
		_, err = os.Stat(filepath.FromSlash(k.state.pwd + "/dict/words.txt"))
		if err != nil {
			k.Log("FATAL", "No dictionary file supplied, please add words.txt to the dict folder, full of words to use for hangman")
			fmt.Println("No dictionary file supplied, please add words.txt to the directory.")
			return
		}
	} else {
		_, err = os.Stat(filepath.FromSlash("/usr/share/dict/words"))
		if err != nil {
			k.Log("FATAL", "Error accessing dictionary at /usr/share/dict/words")
			fmt.Println("Error accesssing dictionary at /usr/share/dict/words.  Try installing the wamerican package.")
			return
		}
	}

	// Create a new Discord session using the provided bot token.
	ky, err := discordgo.New("Bot " + k.botConfig.APIKey)
	if err != nil {
		k.Log("FATAL", "Error creating Discord session")
		panic(err)
	}

	// Keep session globally accessible
	k.session = ky

	// Use the state to cache messages
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

	// BootLogo
	if k.botConfig.BootLogo != "" {
		fmt.Println(k.botConfig.BootLogo)
	}

	// Open the websocket and begin listening for above events.
	err = ky.Open()
	if err != nil {
		fmt.Println("Error opening Discord session")
		panic(err)
	}
	k.Log("STARTUP", "Successfully opened Discord bot session!")

	// Create channels to watch for kill signals
	botChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// Bot will end on any of the following signals
	signal.Notify(botChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	go func() {
		signalType := <-botChan
		k.Log("SHUTDOWN", fmt.Sprintf("Shutting down from signal: %s", signalType))
		done <- true
	}()

	// Wait here until CTRL-C or other term signal is received.
	<-done

	// Cleanly close down the Discord session by disconnecting
	// from any connected voice channels
	k.Log("SHUTDOWN", "Disconnecting from voice channels...")
	if k.state.currentVC.ChannelID != "" {
		k.state.currentVC.Disconnect()
	}
	k.Log("SHUTDOWN", "Closing discord websocket...")
	ky.Close()
	k.Log("SHUTDOWN", "Closing database connection...")
	k.db.Close()
	k.logfile.Close()
	fmt.Println("\nKylixor Bot is now stopped")
	os.Exit(0)
}

// Ready - This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func Ready(s *discordgo.Session, event *discordgo.Ready) {
	k.state.self = event.User

	servers := s.State.Guilds
	k.Log("STARTUP", fmt.Sprintf("Kylixor Discord bot has started on %d servers", len(servers)))
	for _, server := range servers {
		k.kdb.UpdateGuild(s, server.ID)
	}

	// Start cronjobs
	k.cron = cron.New()
	k.cron.AddFunc("0 20 * * *", func() { ResetDailies(s) })
	k.cron.Start()

	// Set status once at start, then ticker takes over every hour
	SetStatus(s)

	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			SetStatus(s)
		}
	}()

	k.Log("STARTUP", "Bot has started...listening for input")

	fmt.Println("\nKylixor Discord bot is now running.  Press CTRL-C to exit.")
}

// MessageReactionAdd - Called whenever a reaction is added to a message
func MessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == k.state.self.ID {
		return
	}

	found, msgType := k.kdb.ReadWatch(r.MessageID)
	if found {
		switch msgType {

		case "hangman":
			// Guess by reaction letter
			hm := k.kdb.ReadHM(s, r.GuildID)
			hm.ReactionGuess(s, r)

		case "vote":
			vote := k.kdb.ReadVote(r.MessageID)
			vote.HandleVote(s, r)
		default:
			return
		}
	}
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
	// Return if the message was sent by a bot to avoid infinite loops
	if m.Author.Bot {
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
		guild := k.kdb.ReadGuild(s, m.GuildID)
		guild.UpdateKarma(s, 1)
		s.MessageReactionAdd(m.ChannelID, m.ID, "😊")
	}

	// Bad Karma
	if strings.ToLower(m.Content) == "bad bot" {
		guild := k.kdb.ReadGuild(s, m.GuildID)
		guild.UpdateKarma(s, -1)
		s.MessageReactionAdd(m.ChannelID, m.ID, "😞")
	}

	if m.Content == fmt.Sprintf("<@%s>", k.state.self.ID) {
		err := s.MessageReactionAdd(m.ChannelID, m.ID, "👋")
		if err != nil {
			panic(err)
		}
	}

	// If the message sent is a command with the set prefix
	if strings.HasPrefix(m.Content, k.botConfig.Prefix) {
		// Trim the prefix to extract the command
		input := strings.TrimPrefix(m.Content, k.botConfig.Prefix)
		//If it is a lone '!', ignore
		if strings.HasPrefix(input, " ") || input == "" {
			return
		}
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

	found, quoteID := k.kdb.ReadWatch(m.ChannelID + m.Author.ID)
	if found {
		if !strings.Contains(m.Content, " ") && len(m.Content) < 20 {
			quote := k.kdb.ReadQuote(s, m.GuildID, quoteID)
			quote.UpdateID(s, m.Content)
			quote.Print(s, m.ChannelID)
			k.kdb.DeleteWatch(m.ChannelID + m.Author.ID)
		} else {
			s.ChannelMessageSend(m.ChannelID, "Identifier must be less than 20 characters and not contain any spaces, try again")
		}
	}
}

//----- I N I T I A L   S E T U P   F U N C T I O N S -----

// Init - Initialize config file if one is not found
func (bc *BotConfig) Init() {
	// Create and indent proper json output for the config
	configData, err := json.MarshalIndent(bc, "", "    ")
	if err != nil {
		panic(err)
	}

	// Open file
	jsonFile, err := os.Create(filepath.FromSlash(k.state.pwd + "/conf.json"))
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
func (bc *BotConfig) Read() {
	file, _ := os.Open(filepath.FromSlash(k.state.pwd + "/conf.json"))
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&bc)
	if err != nil {
		panic(err)
	}
	file.Close()
}

// WriteBotConfig - Write out the current Config structure to file, indented nicely
func (bc *BotConfig) Write() {
	// Indent so its readable
	configData, err := json.MarshalIndent(bc, "", "    ")
	if err != nil {
		panic(err)
	}
	// Open file
	jsonFile, err := os.Create(filepath.FromSlash(k.state.pwd + "/conf.json"))
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
func (bc *BotConfig) Update() {
	bc.Read()
	bc.Write()
}

//----- M I S C .   F U N C T I O N S -----

// GetBotVersion - Get the version of the bot from the readme
func GetBotVersion() (ver string) {
	// Open the file and grab it line by line into textlines
	changelog, err := os.Open(filepath.FromSlash(k.state.pwd + "/CHANGELOG.md"))
	if err != nil {
		panic(err)
	}
	defer changelog.Close()

	scanner := bufio.NewScanner(changelog)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		// Find version number between 2 square brackets
		re := regexp.MustCompile(`\[([^\[\]]*)\]`)
		if re.MatchString(scanner.Text()) {
			versionSplice := re.FindAllString(scanner.Text(), 1)
			ver = versionSplice[0]
			ver = strings.Trim(ver, "[")
			ver = strings.Trim(ver, "]")
			k.botConfig.Version = ver
			return ver
		}

	}
	return "ERR"
}

// SetStatus - sets the status of the bot to the version and the default help commands
func SetStatus(s *discordgo.Session) {
	s.UpdateStatus(0, fmt.Sprintf("%shelp - v%s", k.botConfig.Prefix, k.botConfig.Version))
}

// CheckAdmin - returns true if user is admin, otherwise posts that permission is denied
func CheckAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if k.botConfig.Admin == m.Author.ID {
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

// GetAge - Take in a string that should be a discord snowflake ID, then calculate, format, and return the age
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

// Print (quote) - Prints the quote with nice colors
func (quote *Quote) Print(s *discordgo.Session, cID string) *discordgo.Message {
	// Format quote with colors using the CSS formatting
	fmtQuote := fmt.Sprintf("```ini\n[ %s ] - [ %s ]\n%s\n```", quote.Timestamp.Format("Jan 2 3:04:05PM 2006"), quote.Identifier, quote.Quote)
	// Return the sent message for vote monitoring
	msg, err := s.ChannelMessageSend(cID, fmtQuote)
	if err != nil {
		panic(err)
	}
	return msg
}
