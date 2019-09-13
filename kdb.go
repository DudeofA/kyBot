/* 	kdb.go
_________________________________
Manipulating 'database' for guild data for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//----- M O N G O   C O N N E C T I O N -----
var mdbClient *mongo.Client

//----- K D B   S T R U C T U R E -----
var kdb KDB

// KDB - Structure for holding pointers to all necessary data
type KDB struct {
	// Collections within database
	UserColl     *mongo.Collection // Collection for user info
	GuildColl    *mongo.Collection // Collection for guilds/server info
	HangmanColl  *mongo.Collection // Collection for hangman games
	QuoteColl    *mongo.Collection // Collection for all quotes
	ReminderColl *mongo.Collection // Collection for all reminders
}

//----- S E R V E R   I N F O -----

// GuildInfo - Hold all the pertaining information for each server
type GuildInfo struct {
	Config GuildConfig `json:"config" bson:"serverConfig"` // Guild specific config
	Emotes Emote       `json:"emotes" bson:"emotes"`       // String of customizable emotes
	ID     string      `json:"gID" bson:"gID"`             // discord guild ID
	Karma  int         `json:"karma" bson:"karma"`         // Bot's karma - per guild
	Name   string      `json:"name" bson:"name"`           // Name of guild
	Region string      `json:"region" bson:"region"`       // Geolocation region of the guild
}

// GuildConfig - structure to hold variables specifically for that guild
type GuildConfig struct {
	Coins    string `json:"coins" bson:"coins"`       // Name of currency that bot uses (i.e. <gold> coins)
	Follow   bool   `json:"follow" bson:"follow"`     // Whether or not the bot joins/follows into voice channels for anthems
	MinVotes int    `json:"minVotes" bson:"minVotes"` // Minimum upvotes to pass a vote
	Prefix   string `json:"prefix" bson:"prefix"`     // Prefix the bot will respond to
}

// Emote - customizable emotes for reactions the bot adds
type Emote struct {
	Upvote   string `json:"upvote" bson:"upvote"`     // Upvote emotes
	Downvote string `json:"downvote" bson:"downvote"` // Downvote emotes
}

// Hangman - State of hangman game
type Hangman struct {
	GuildID   string   `json:"guildID" bson:"guildID"`     // Guild game is attached to
	Channel   string   `json:"channel" bson:"channel"`     // ChannelID where game is played
	GameState int      `json:"gameState" bson:"gameState"` // State of game, 1-7 until you lose
	Guessed   []string `json:"guessed" bson:"guessed"`     // Characters/words that have been guessed
	Message   string   `json:"message" bson:"message"`     // MessageID of current hangman game
	Word      string   `json:"word" bson:"word"`           // Word/phrase for the game
	WordState []string `json:"hmState" bson:"hmState"`     // State of game's word
}

// Quote - Data about quotes and quotes themselves
type Quote struct {
	GuildID    string    `json:"guildID" bson:"guildID"`       // Guild quote is from
	Identifier string    `json:"identifier" bson:"identifier"` // Word to identify quote
	Quote      string    `json:"quote" bson:"quote"`           // Actual quoted text
	Timestamp  time.Time `json:"timestamp" bson:"timestamp"`   // Timestamp when quote was recorded
}

//----- U S E R   S T A T S -----

// UserInfo - Hold all pertaining information for each user
type UserInfo struct {
	ID            string `json:"userID" bson:"userID"`               // User ID
	Name          string `json:"name" bson:"name"`                   // Username
	Discriminator string `json:"discriminator" bson:"discriminator"` // Unique identifier (#4712)
	CurrentCID    string `json:"currentCID" bson:"currentCID"`       // Current channel ID
	LastSeenCID   string `json:"lastSeenCID" bson:"lastSeenCID"`     // Last seen channel ID
	Credits       int    `json:"credits" bson:"credits"`             // Credits gained from dailies
	DoneDailies   bool   `json:"dailies" bson:"dailies"`             // True if dailies have been claimed today
}

// Reminders - holds reminders for the bot to tell the user about
type Reminders struct {
	UserID     string    `json:"userID" bson:"userID"`         // User that saved the reminder
	RemindTime time.Time `json:"remindTime" bson:"remindTime"` // Time to remind user
	RemindMsg  string    `json:"remindMsg" bson:"remindMsg"`   // Message to be reminded of
}

//----- M O N G O D B   F U N C T I O N S -----

// Init - attempt to make a connection to the Mongo database
func (k *KDB) Init() {
	fmt.Println("Connecting to MongoDB...")
	// Create client using URI provided
	var err error
	clientOptions := options.Client().ApplyURI(botConfig.DBConfig.URI)
	mdbClient, err = mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Println("Error creating MongoDB client")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Connect to MongoDB
	err = mdbClient.Connect(context.Background())
	if err != nil {
		fmt.Println("Error connecting to the database")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Define collections
	mdb := mdbClient.Database(botConfig.DBConfig.DBName)
	k.GuildColl = mdb.Collection("guilds")
	k.HangmanColl = mdb.Collection("hangmanGames")
	k.QuoteColl = mdb.Collection("quotes")
	k.ReminderColl = mdb.Collection("reminders")
	k.UserColl = mdb.Collection("users")
	fmt.Println("Connected to MongoDB!")
}

//----- G U I L D   M A N A G E M E N T -----

// CreateGuild - Create server from given ID
func (k *KDB) CreateGuild(s *discordgo.Session, id string) (guild GuildInfo) {
	// Get guild info from Discord
	discordGuild, err := s.Guild(id)
	if err != nil {
		LogTxt(s, "FATAL", "Invalid guild ID passed to CreateGuild")
		panic(err)
	}

	guild.ID = id
	guild.Name = discordGuild.Name
	guild.Region = discordGuild.Region
	// Set emotes to default
	guild.Emotes.Upvote = "⬆"
	guild.Emotes.Downvote = "⬇"
	// Set prefix to default
	guild.Config.Prefix = "k!"
	// Set votes to default minimum
	guild.Config.MinVotes = 3

	// Add the new guild to the guild collection
	objID, err := k.GuildColl.InsertOne(context.Background(), guild)
	if err != nil {
		panic(err)
	}

	LogDB(s, "Guild", guild.Name, guild.ID, "created", objID.InsertedID)

	return guild
}

// ReadGuild - Get the server information from the db
func (k *KDB) ReadGuild(s *discordgo.Session, id string) (guild GuildInfo) {

	filter := bson.D{{"gID", id}}
	err := k.GuildColl.FindOne(context.Background(), filter).Decode(&guild)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return k.CreateGuild(s, id)
		}
		panic(err)
	}

	return guild
}

// Update [Guild] - update guild in database based on argument
func (guild *GuildInfo) Update(s *discordgo.Session) {
	filter := bson.D{{"guildID", guild.ID}}
	result := kdb.HangmanColl.FindOneAndReplace(context.Background(), filter, guild)
	if result.Err() != nil {
		LogTxt(s, "ERR", "Guild was not modified")
		panic(result.Err())
	}

	LogDB(s, "Guild", guild.Name, guild.ID, "updated", "N/A")

}

//----- U S E R   M A N A G E M E N T -----

// CreateUser - create user with default values and return it
func (k *KDB) CreateUser(s *discordgo.Session, id string) (user UserInfo) {

	// Get user info from discord
	discordUser, err := s.User(id)
	if err != nil {
		panic(err)
	}

	user.ID = id
	user.Name = discordUser.Username
	user.Discriminator = discordUser.Discriminator
	// Set defaults
	user.DoneDailies = false

	// Insert user into collection
	objID, err := kdb.UserColl.InsertOne(context.Background(), user)
	if err != nil {
		panic(err)
	}

	LogDB(s, "User", user.Name, user.ID, "created", objID.InsertedID)

	return user
}

// ReadUser - Query database for user, creating a new one if none exists
func (k *KDB) ReadUser(s *discordgo.Session, userID string) (user UserInfo) {

	// Search by discord user ID
	filter := bson.D{{"userID", userID}}
	err := k.UserColl.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return k.CreateUser(s, userID)
		}
		panic(err)
	}

	return user
}

// Update [User] - update user in database based on user argument
func (user *UserInfo) Update(s *discordgo.Session) {

	// Search by discord user ID
	filter := bson.D{{"userID", user.ID}}
	result := kdb.UserColl.FindOneAndReplace(context.Background(), filter, user, {upsert: true})
	if result.Err() != nil {
		LogTxt(s, "ERR", "User was not modified")
		panic(result.Err())
	}

	LogDB(s, "User", user.Name, user.ID, "updated", "N/A")

}

//----- H A N G M A N   M A N A G E M E N T -----

// CreateHM - insert the hangman game into the hangman collection
func (k *KDB) CreateHM(s *discordgo.Session, guildID string) (hm Hangman) {

	hm.GuildID = guildID
	hm.ResetGame()

	// Insert hangman game into collection
	objID, err := kdb.HangmanColl.InsertOne(context.Background(), hm)
	if err != nil {
		panic(err)
	}

	LogDB(s, "Hangman Game", hm.Word, hm.GuildID, "created", objID.InsertedID)

	return hm
}

// ReadHM - get hangman session from hangman collection
func (k *KDB) ReadHM(s *discordgo.Session, guildID string) (hm Hangman) {

	// Search by discord guild ID
	filter := bson.D{{"guildID", guildID}}
	err := k.HangmanColl.FindOne(context.Background(), filter).Decode(&hm)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return k.CreateHM(s, guildID)
		}
		panic(err)
	}

	return hm
}

// Update [HM] - update hangman game in the database
func (hm *Hangman) Update(s *discordgo.Session) {
	filter := bson.D{{"guildID", hm.GuildID}}
	result := kdb.HangmanColl.FindOneAndReplace(context.Background(), filter, hm)
	if result.Err() != nil {
		LogTxt(s, "ERR", "Hangman game was not modified")
		panic(result.Err())
	}

	LogDB(s, "Hangman", hm.Word, hm.GuildID, "updated", "N/A")

}

//----- Q U O T E   M A N A G E M E N T -----

// CreateQuote - create quote and insert it into the database
func (k *KDB) CreateQuote(s *discordgo.Session, guildID string, quoteText string) (quote Quote) {

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
	objID, err := kdb.QuoteColl.InsertOne(context.Background(), quote)
	if err != nil {
		panic(err)
	}

	LogDB(s, "Quote", quote.Identifier, quote.GuildID, "inserted", objID.InsertedID)

	return quote
}

// ReadQuote - try to get the vote from the database, returning empty quote if none found
//	returns random quote with no argument
func (k *KDB) ReadQuote(guildID, identifier string) (quote Quote) {

	// Search by discord guild ID & identifier
	filter := bson.D{{"guildID", guildID}, {"identifier", identifier}}
	err := k.QuoteColl.FindOne(context.Background(), filter).Decode(&quote)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return quote
		}
		panic(err)
	}

	return quote
}

// Update [Quote] - update quote in database
func (quote *Quote) Update(s *discordgo.Session) {
	filter := bson.D{{"guildID", quote.GuildID}, {"identifier", quote.Identifier}}
	result := kdb.HangmanColl.FindOneAndReplace(context.Background(), filter, quote)
	if result.Err() != nil {
		LogTxt(s, "ERR", "Quote was not modified")
		panic(result.Err())
	}

	LogDB(s, "Quote", quote.Identifier, quote.GuildID, "updated", "N/A")

}
