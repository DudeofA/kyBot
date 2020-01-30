/* 	kdb.go
_________________________________
Manipulating 'database' for guild data for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"database/sql"
	"fmt"
	"math/rand"
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

	// Create state table
	k.Log("KDB", "Creating state table and updating version")
	_, err = k.db.Exec(`CREATE TABLE IF NOT EXISTS state (
		version VARCHAR(32) NOT NULL)`)
	if err != nil {
		panic(err)
	}

	_, err = k.db.Exec(fmt.Sprintf("INSERT INTO state (version) VALUES ('%s')",
		k.botConfig.Version))
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

	row := k.db.QueryRow("SELECT guildID, name, region, karma, currency FROM guilds WHERE guildID=(?)", guildID)
	err := row.Scan(&guild.ID, &guild.Name, &guild.Region, &guild.Karma, &guild.Currency)
	switch err {
	case sql.ErrNoRows:
		k.Log("KDB", "Guild not found in DB, creating new...")
		return k.kdb.CreateGuild(s, guildID)
	case nil:
		LogDB("Guild", guild.Name, guild.ID, "read from")
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

//----- Q U O T E   M A N A G E M E N T -----

// CreateQuote - create quote and insert it into the database
func (kdb *KDB) CreateQuote(s *discordgo.Session, guildID string, quoteText string) (quote Quote) {

	// Generate 3 letter identifier
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, 3)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	identifier := string(b)

	// Add quote to quote object
	quote = Quote{guildID, identifier, quoteText, time.Now()}

	// Add quote to quote collection
	_, err := k.db.Exec("INSERT INTO quotes (identifier, guildID, quote, timestamp) VALUES(?,?,?,?)",
		quote.Identifier, quote.GuildID, quote.Quote, quote.Timestamp.Format("2006-01-02 15:04:05"))
	if err != nil {
		panic(err)
	}

	LogDB("Quote", quote.Identifier, quote.GuildID, "inserted into")

	return quote
}

// ReadQuote - try to get the vote from the database, returning empty quote if none found
//	returns random quote with no argument
func (kdb *KDB) ReadQuote(s *discordgo.Session, guildID, identifier string) (quote Quote) {
	var tempTime string

	if identifier == "" {
		rand := k.db.QueryRow("SELECT identifier, guildID, quote, timestamp FROM quotes ORDER BY RAND()")
		err := rand.Scan(&quote.Identifier, &quote.GuildID, &quote.Quote, &tempTime)
		switch err {
		case sql.ErrNoRows:
			LogDB("Quote", quote.Identifier, quote.GuildID, "not found in")
			return Quote{}
		case nil:
			LogDB("Quote", quote.Identifier, quote.GuildID, "read from")
			quote.Timestamp, err = time.Parse("2006-01-02 15:04:05", tempTime)
			if err != nil {
				panic(err)
			}
			return quote
		default:
			panic(err)
		}
	}

	// Search by discord guild ID & identifier
	row := k.db.QueryRow("SELECT identifier, guildID, quote, timestamp FROM quotes WHERE guildID=(?) AND identifier=(?)", guildID, identifier)
	err := row.Scan(&quote.Identifier, &quote.GuildID, &quote.Quote, &tempTime)
	switch err {
	case sql.ErrNoRows:
		LogDB("Quote", quote.Identifier, quote.GuildID, "not found")
		return Quote{}
	case nil:
		LogDB("Quote", quote.Identifier, quote.GuildID, "read from")
		quote.Timestamp, err = time.Parse("2006-01-02 15:04:05", tempTime)
		if err != nil {
			panic(err)
		}
		return quote
	default:
		panic(err)
	}
}

// Update [Quote] - update quote in database
func (quote *Quote) Update(s *discordgo.Session) {

	LogDB("Quote", quote.Identifier, quote.GuildID, "updated in")

	//TODO

}

//----- G U I L D   S T A T   M A N A G E M E N T -----

// UpdateKarma - change the karma of the server
func (guild GuildInfo) UpdateKarma(s *discordgo.Session, delta int) {

	LogDB("KARMA", guild.Name, guild.ID, "changed by "+strconv.Itoa(delta))

	guild.Karma += delta
	_, err := k.db.Exec("UPDATE guilds SET karma = karma+? WHERE guildID=?", delta, guild.ID)
	if err != nil {
		panic(err)
	}
}
