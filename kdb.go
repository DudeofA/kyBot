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
	"os"
	"time"

	"github.com/bwmarrin/discordgo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//----- M A S T E R   D A T A B A S E
// Kylixor """"Database""""
var kdb KDB

//----- K D B   S T R U C T U R E -----

// KDB - Structure for holding pointers to all necessary data
type KDB struct {
	DB     *mongo.Database // Mongo database pointer
	Client *mongo.Client   // Database client connection

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
	UPVOTE   string `json:"upvote" bson:"upvote"`     // Upvote emotes
	DOWNVOTE string `json:"downvote" bson:"downvote"` // Downvote emotes
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
	GuildID   string    `json:"guildID" bson:"guildID"`     // Guild quote is from
	Quote     string    `json:"quote" bson:"quote"`         // Actual quoted text
	Timestamp time.Time `json:"timestamp" bson:"timestamp"` // Timestamp when quote was recorded
}

//----- U S E R   S T A T S -----

// UserInfo - Hold all pertaining information for each user
type UserInfo struct {
	ID            string   `json:"userID" bson:"userID"`               // User ID
	Name          string   `json:"name" bson:"name"`                   // Username
	Discriminator string   `json:"discriminator" bson:"discriminator"` // Unique identifier
	Guilds        []string `json:"guilds" bson:"guilds"`               // List of Guild IDs user is a part of
	CurrentCID    string   `json:"currentCID" bson:"currentCID"`       // Current channel ID
	LastSeenCID   string   `json:"lastSeenCID" bson:"lastSeenCID"`     // Last seen channel ID
	PlayAnthem    bool     `json:"playAnthem" bson:"playAnthem"`       // True if anthem should play when user joins channel
	Anthem        string   `json:"anthem" bson:"anthem"`               // Anthem to play when joining a channel
	Credits       int      `json:"credits" bson:"credits"`             // Credits gained from dailies
	DoneDailies   bool     `json:"dailies" bson:"dailies"`             // True if dailies have been claimed today
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
	// Client parameters
	var err error
	clientOptions := options.Client().ApplyURI(botConfig.DBConfig.URI)
	kdb.Client, err = mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Println("Error creating MongoDB client")
		os.Exit(1)
	}

	// Connect to MongoDB
	err = kdb.Client.Connect(context.Background())
	if err != nil {
		fmt.Println("Error connecting to the database")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	k.DB = k.Client.Database(botConfig.DBConfig.DBName)
	k.GuildColl = k.DB.Collection("guilds")
	k.HangmanColl = k.DB.Collection("hangmanGames")
	k.QuoteColl = k.DB.Collection("quotes")
	k.ReminderColl = k.DB.Collection("reminders")
	k.UserColl = k.DB.Collection("users")
	fmt.Println("Connected to MongoDB!")
}

//----- G U I L D   M A N A G E M E N T -----

// GetGuild - Get the server information from the db
func (k *KDB) GetGuild(s *discordgo.Session, id string) (guild GuildInfo) {
	filter := bson.D{{"gID", id}}
	err := k.GuildColl.FindOne(context.Background(), filter).Decode(&guild)
	if err != nil {
		return k.CreateGuild(s, id)
	}

	return k.CreateGuild(s, id)
}

// CreateGuild - Create server from given ID
func (k *KDB) CreateGuild(s *discordgo.Session, id string) (guild GuildInfo) {

	discordGuild, err := s.Guild(id)
	if err != nil {
		panic(err)
	}

	// Guild not found - Create new
	guild.ID = id
	guild.Name = discordGuild.Name
	guild.Region = discordGuild.Region
	// Set emotes to default
	guild.Emotes.UPVOTE = "⬆"
	guild.Emotes.DOWNVOTE = "⬇"
	// Set prefix to default
	guild.Config.Prefix = "k!"
	// Set votes to default minimum
	guild.Config.MinVotes = 3

	// Add the new guild
	objID, err := k.GuildColl.InsertOne(context.Background(), guild)
	if err != nil {
		panic(err)
	}

	LogTxt(s, "INFO", fmt.Sprintf("Guild \"%s\" [%s] inserted into DB (MongoID#%S)",
		guild.Name, guild.ID, objID.InsertedID))

	return guild
}

// UpdateGuild - update guild in database based on argument
func (k *KDB) UpdateGuild(guild GuildInfo) {
	//
}

//----- U S E R   M A N A G E M E N T -----

// GetUser - Query database for user, creating a new one if none exists
func (k *KDB) GetUser(s *discordgo.Session, id string) (user UserInfo) {
	filter := bson.D{{"userID", id}}
	err := k.UserColl.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return k.CreateUser(s, id)
	}
	return user
}

// CreateUser - create user with default values and return it
func (k *KDB) CreateUser(s *discordgo.Session, id string) (user UserInfo) {
	// Get user info from discord
	discordUser, err := s.User(id)
	if err != nil {
		panic(err)
	}

	// Set unique values of user
	user.ID = id
	user.Name = discordUser.Username
	user.Discriminator = discordUser.Discriminator
	// Set defaults
	user.PlayAnthem = false
	user.DoneDailies = false

	// Insert user into collection
	objID, err := kdb.UserColl.InsertOne(context.Background(), user)
	if err != nil {
		panic(err)
	}
	LogTxt(s, "INFO", fmt.Sprintf("User \"%s\" [%s] inserted into DB (MongoID#%S)",
		user.Name, user.ID, objID.InsertedID))

	return
}

// UpdateUser - update user in database based on user argument
func (k *KDB) UpdateUser(user UserInfo) {
	//
}

//----- H A N G M A N   M A N A G E M E N T -----

// GetHM - get hangman session from hangman collection
func (k *KDB) GetHM(guildID string) (hm Hangman) {
	//

	return hm
}

// CreateHM - create hangman session if none exists
func (k *KDB) CreateHM(guildID string) (hm Hangman) {
	//

	return hm
}

// UpdateHM - update hangman game in the database
func (k *KDB) UpdateHM(hm Hangman) {
	//
}

//----- Q U O T E   M A N A G E M E N T -----

// CreateQuote - create quote and insert it into the database
func (k *KDB) CreateQuote(quote Quote) {
	//
}
