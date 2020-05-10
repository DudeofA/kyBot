/* 	kdb.go
_________________________________
Manipulating 'database' for guild data for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
)

// KDB - Structure for holding pointers to all necessary data
type KDB struct {
}

//----- S E R V E R   I N F O -----

// GuildInfo - Hold all the pertaining information for each server
type GuildInfo struct {
	Currency string `json:"currency"` // Name of currency that bot uses (i.e. <gold> coins)
	ID       string `json:"guildID"`  // discord guild ID
	Karma    int    `json:"karma"`    // Bot's karma - per guild
	Name     string `json:"name"`     // Name of guild
	Region   string `json:"region"`   // Geolocation region of the guild
}

// Quote - Data about quotes and quotes themselves
type Quote struct {
	GuildID    string    `json:"guildID"`    // Guild quote is from
	Identifier string    `json:"identifier"` // Word to identify quote
	Quote      string    `json:"quote"`      // Actual quoted text
	Timestamp  time.Time `json:"timestamp"`  // Timestamp when quote was recorded
}

// Hangman - State of hangman game
type Hangman struct {
	GuildID   string   `json:"guildID"`   // Guild game is attached to
	ChannelID string   `json:"channelID"` // ChannelID where game is played
	GameState int      `json:"gameState"` // State of game, 1-7 until you lose
	Guessed   []string `json:"guessed"`   // Characters/words that have been guessed
	MessageID string   `json:"messageID"` // MessageID of current hangman game
	Word      string   `json:"word"`      // Word/phrase for the game
	WordState []string `json:"hmState"`   // State of game's word
}

//----- U S E R   S T A T S -----

// UserInfo - Hold all pertaining information for each user
type UserInfo struct {
	ID            string `json:"userID"`        // User ID
	Name          string `json:"name"`          // Username
	Discriminator string `json:"discriminator"` // Unique identifier (#4712)
	CurrentCID    string `json:"currentCID"`    // Current channel ID
	LastSeenCID   string `json:"lastSeenCID"`   // Last seen channel ID
	Credits       int    `json:"credits"`       // Credits gained from dailies
	DoneDailies   bool   `json:"dailies"`       // True if dailies have been claimed today
}

// Reminders - holds reminders for the bot to tell the user about
type Reminders struct {
	UserID     string    `json:"userID"`     // User that saved the reminder
	RemindTime time.Time `json:"remindTime"` // Time to remind user
	RemindMsg  string    `json:"remindMsg"`  // Message to be reminded of
}

// Vote - hold stats for votes
type Vote struct {
	MessageID   string    `json:"messageID"`   // MessageID of vote
	GuildID     string    `json:"guildID"`     // Guild vote is in
	SubmitterID string    `json:"submitterID"` // ID of submitter
	Options     int       `json:"options"`     // How many options the vote has
	Quote       bool      `json:"quote"`       // Is vote a quote application
	VoteText    string    `json:"voteText"`    // All the options for the vote in one string
	Result      int       `json:"result"`      // Numeric result of vote (negative is in progress, zero is no for single option vote, numbers are voting options)
	StartTime   time.Time `json:"startTime"`   // Time vote was started
	EndTime     time.Time `json:"endTime"`     // Time vote was ended, by default it is same as startTime
}

//----- D B   F U N C T I O N S -----

// Init - setup tables
func (kdb *KDB) Init() {
	var err error

	// Create users table
	k.Log("KDB", "Creating users table")
	_, err = k.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		userID VARCHAR(32) PRIMARY KEY NOT NULL, 
		name VARCHAR(32) NOT NULL, 
		discriminator VARCHAR(4) NOT NULL, 
		currentCID VARCHAR(32) NOT NULL DEFAULT '', 
		lastSeenCID VARCHAR(32) NOT NULL DEFAULT '', 
		credits INT NOT NULL DEFAULT 0, 
		dailies BOOL NOT NULL DEFAULT FALSE)`)
	if err != nil {
		panic(err)
	}

	// Create guilds table
	k.Log("KDB", "Creating guild table")
	_, err = k.db.Exec(`CREATE TABLE IF NOT EXISTS guilds (
		guildID VARCHAR(32) PRIMARY KEY NOT NULL,
		name VARCHAR(32) NOT NULL,
		region VARCHAR(32) NOT NULL,
		karma INT NOT NULL DEFAULT 0,
		currency VARCHAR(32) NOT NULL DEFAULT '')`)
	if err != nil {
		panic(err)
	}

	// Create hangman table
	k.Log("KDB", "Creating hangman table")
	_, err = k.db.Exec(`CREATE TABLE IF NOT EXISTS hangman (
		guildID VARCHAR(32) PRIMARY KEY NOT NULL,
		channelID VARCHAR(32) NOT NULL,
		messageID VARCHAR(32) NOT NULL,
		word VARCHAR(32) NOT NULL,
		gameState INT NOT NULL,
		wordState VARCHAR(32) NOT NULL,
		guessed TEXT NOT NULL)`)
	if err != nil {
		panic(err)
	}

	// Create quote table
	k.Log("KDB", "Creating quote table")
	_, err = k.db.Exec(`CREATE TABLE IF NOT EXISTS quotes (
		identifier VARCHAR(32) PRIMARY KEY NOT NULL,
		guildID VARCHAR(32) NOT NULL,
		quote TEXT NOT NULL,
		timestamp DATETIME NOT NULL)`)
	if err != nil {
		panic(err)
	}

	// Create vote table
	k.Log("KDB", "Creating votes table")
	_, err = k.db.Exec(`CREATE TABLE IF NOT EXISTS votes (
		messageID VARCHAR(32) PRIMARY KEY NOT NULL,
		guildID VARCHAR(32) NOT NULL,
		submitterID VARCHAR(32) NOT NULL,
		options INT NOT NULL,
		quote BOOL NOT NULL DEFAULT FALSE,
		voteText TEXT NOT NULL,
		result INT NOT NULL DEFAULT 0,
		startTime DATETIME NOT NULL,
		endTime DATETIME NOT NULL)`)
	if err != nil {
		panic(err)
	}

	// Create watch table
	k.Log("KDB", "Creating watch table")
	_, err = k.db.Exec(`CREATE TABLE IF NOT EXISTS watch (
		messageID VARCHAR(32) PRIMARY KEY NOT NULL,
		type VARCHAR(32) NOT NULL)`)
	if err != nil {
		panic(err)
	}

	// Create state table
	k.Log("KDB", "Creating state table and updating version")
	_, err = k.db.Exec(`CREATE TABLE IF NOT EXISTS state (
		version VARCHAR(32) NOT NULL)`)
	if err != nil {
		panic(err)
	}

	_, err = k.db.Exec("INSERT INTO state (version) VALUES (?)", k.botConfig.Version)
	if err != nil {
		panic(err)
	}
}

//----- G U I L D   M A N A G E M E N T -----

// CreateGuild - Create server from given ID
func (kdb *KDB) CreateGuild(s *discordgo.Session, id string) (guild GuildInfo) {

	// Get guild info from Discord
	discordGuild, err := s.Guild(id)
	if err != nil {
		k.Log("FATAL", "Invalid guild ID passed to CreateGuild")
		panic(err)
	}

	guild.ID = id
	guild.Name = discordGuild.Name
	guild.Region = discordGuild.Region

	// Add the new guild to the database
	_, err = k.db.Exec("INSERT INTO guilds (guildID, name, region, currency) VALUES(?,?,?,?)",
		guild.ID, guild.Name, guild.Region, "")
	if err != nil {
		panic(err)
	}

	LogDB("Guild", guild.Name, guild.ID, "created in")
	return guild
}

// ReadGuild - Get the server information from the db
func (kdb *KDB) ReadGuild(s *discordgo.Session, guildID string) (guild GuildInfo) {
	// TODO - IF CHAT IS GOING THROUGH DIRECT MESSAGE
	// if id == "" {
	// 	return discordgo.Guild {
	// 		Name = "Direct Message",
	// 		ID = s.UserChannelCreate()
	// 	}
	// }

	row := k.db.QueryRow("SELECT guildID, name, region, karma, currency FROM guilds WHERE guildID = ?", guildID)
	err := row.Scan(&guild.ID, &guild.Name, &guild.Region, &guild.Karma, &guild.Currency)
	switch err {
	case sql.ErrNoRows:
		k.Log("KDB", "Guild not found in DB, creating new...")
		return k.kdb.CreateGuild(s, guildID)
	case nil:
		// LogDB("Guild", guild.Name, guild.ID, "read from")
		return guild
	default:
		panic(err)
	}
}

// UpdateGuild - update guild in database based on argument
func (kdb *KDB) UpdateGuild(s *discordgo.Session, id string) {
	// Get guild info from Discord
	guild, err := s.Guild(id)
	if err != nil {
		k.Log("FATAL", "Invalid guild ID passed to UpdateGuild")
		panic(err)
	}

	LogDB("Guild", guild.Name, guild.ID, "updated in")

	// Update the guild to the database
	_, err = k.db.Exec("INSERT INTO guilds (guildID, name, region) VALUES(?,?,?) ON DUPLICATE KEY UPDATE name = ?, region = ?",
		guild.ID, guild.Name, guild.Region, guild.Name, guild.Region)
	if err != nil {
		k.Log("FATAL", "Error updating guild: "+guild.Name)
		panic(err)
	}
}

//----- G U I L D   S T A T   M A N A G E M E N T -----

// UpdateKarma - change the karma of the server
func (guild GuildInfo) UpdateKarma(s *discordgo.Session, delta int) {

	LogDB("KARMA", guild.Name, guild.ID, "changed by "+strconv.Itoa(delta))

	guild.Karma += delta
	_, err := k.db.Exec("UPDATE guilds SET karma = karma+? WHERE guildID = ?", delta, guild.ID)
	if err != nil {
		panic(err)
	}
}

//----- W A T C H   M A N A G E M E N T -----

// CreateWatch - Start monitoring a message for reactions
func (kdb *KDB) CreateWatch(messageID string, watchType string) {
	_, err := k.db.Exec("INSERT INTO watch (messageID, type) VALUES (?,?)", messageID, watchType)
	if err != nil {
		panic(err)
	}
}

// ReadWatch - Checks watch table for the messageID passed as argument
func (kdb *KDB) ReadWatch(messageID string) (inTable bool, watchType string) {
	row := k.db.QueryRow("SELECT type FROM watch WHERE messageID = ?", messageID)
	err := row.Scan(&watchType)
	if err == sql.ErrNoRows {
		return false, ""
	} else if err == nil {
		return true, watchType
	} else {
		panic(err)
	}
}

// DeleteWatch - Remove watch from watch table based on passed messageID
func (kdb *KDB) DeleteWatch(messageID string) {
	_, err := k.db.Exec("DELETE FROM watch WHERE messageID=?", messageID)
	if err == sql.ErrNoRows {
		k.Log("HANGMAN", "Unable to delete watch table entry, not found")
	} else if err != nil {
		panic(err)
	}
}

//----- V O T E   M A N A G E M E N T -----

// CreateVote - Create a vote and store it in the votes table
func (kdb *KDB) CreateVote(messageID string, guildID string, submitterID string, options int, quote bool, voteText string) (vote Vote) {

	vote.MessageID = messageID
	vote.GuildID = guildID
	vote.SubmitterID = submitterID
	vote.Options = options
	vote.Quote = quote
	vote.VoteText = voteText
	vote.Result = -1
	vote.StartTime = time.Now()
	vote.EndTime = time.Now()
	startTime := vote.StartTime.Format("2006-01-02 15:04:05")
	endTime := vote.EndTime.Format("2006-01-02 15:04:05")

	_, err := k.db.Exec("INSERT INTO votes (messageID, guildID, options, quote, voteText, result, startTime, endTime) VALUES (?,?,?,?,?,?,?,?)",
		vote.MessageID, vote.GuildID, vote.Options, vote.Quote, vote.VoteText, vote.Result, startTime, endTime)
	if err != nil {
		panic(err)
	}

	LogDB("VOTE", guildID, messageID, "created in")
	return vote
}

// ReadVote - Return the vote structure from the database if it exists
func (kdb *KDB) ReadVote(messageID string) (vote Vote) {
	var startTime, endTime string
	row := k.db.QueryRow("SELECT messageID, guildID, submitterID, options, quote, voteText, result, startTime, endTime FROM votes WHERE messageID = ?", messageID)
	err := row.Scan(&vote.MessageID, &vote.GuildID, &vote.SubmitterID, &vote.Options, &vote.Quote, &vote.VoteText, &vote.Result, &startTime, &endTime)
	if err == sql.ErrNoRows {
		return Vote{}
	} else if err == nil {
		vote.StartTime, err = time.Parse("2006-01-02 15:04:05", startTime)
		if err != nil {
			panic(err)
		}
		vote.EndTime, err = time.Parse("2006-01-02 15:04:05", endTime)
		if err != nil {
			panic(err)
		}
		LogDB("VOTE", vote.GuildID, vote.MessageID, "read from")
		return vote
	} else {
		panic(err)
	}
}

// UpdateVote - Update the results of a vote
func (vote *Vote) UpdateVote() {
	// Update the guild to the database
	_, err := k.db.Exec("UPDATE votes SET result = ? WHERE messageID = ?", vote.Result, vote.MessageID)
	if err != nil {
		k.Log("FATAL", "Error updating vote result: "+vote.MessageID)
		panic(err)
	}
}

// EndVote - Put the current time in for the end of the vote
func (vote *Vote) EndVote() {
	vote.EndTime = time.Now()
	endTime := vote.EndTime.Format("2006-01-02 15:04:05")

	_, err := k.db.Exec("UPDATE votes SET endTime = ? WHERE messageID = ?", endTime, vote.MessageID)
	if err != nil {
		k.Log("FATAL", "Error ending vote: "+vote.MessageID)
		panic(err)
	}
}
