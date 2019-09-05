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
	UserColl     *mongo.Collection
	GuildColl    *mongo.Collection
	ReminderColl *mongo.Collection
}

//----- S E R V E R   I N F O -----

// GuildInfo - Hold all the pertaining information for each server
type GuildInfo struct {
	Config GuildConfig `json:"config" bson:"serverConfig"` // Guild specific config
	Emotes Emote       `json:"emotes" bson:"emotes"`       // String of customizable emotes
	GID    string      `json:"gID" bson:"gID"`             // discord guild ID
	HM     Hangman     `json:"hangman" bson:"hangman"`     // Holds hangman data
	Karma  int         `json:"karma" bson:"karma"`         // Bot's karma - per server
	Name   string      `json:"name" bson:"name"`           // Name of server
	Quotes []Quote     `json:"quotes" bson:"quotes"`       // Array of quotes
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
	Channel   string   `json:"channel" bson:"channel"`     // ChannelID where game is played
	GameState int      `json:"gameState" bson:"gameState"` // State of game, 1-7 until you lose
	Guessed   []string `json:"guessed" bson:"guessed"`     // Characters/words that have been guessed
	Message   string   `json:"message" bson:"message"`     // MessageID of current hangman game
	Word      string   `json:"word" bson:"word"`           // Word/phrase for the game
	WordState []string `json:"hmState" bson:"hmState"`     // State of game's word
}

// Quote - Data about quotes and quotes themselves
type Quote struct {
	Quote     string    `json:"quote" bson:"quote"`         // Actual quoted text
	Timestamp time.Time `json:"timestamp" bson:"timestamp"` // Timestamp when quote was recorded
}

//----- U S E R   S T A T S -----

// UserInfo - Hold all pertaining information for each user
type UserInfo struct {
	UserID        string `json:"userID" bson:"userID"`               // User ID
	Name          string `json:"name" bson:"name"`                   // Username
	Discriminator string `json:"discriminator" bson:"discriminator"` // Unique identifier
	CurrentCID    string `json:"currentCID" bson:"currentCID"`       // Current channel ID
	LastSeenCID   string `json:"lastSeenCID" bson:"lastSeenCID"`     // Last seen channel ID
	PlayAnthem    bool   `json:"playAnthem" bson:"playAnthem"`       // True if anthem should play when user joins channel
	Anthem        string `json:"anthem" bson:"anthem"`               // Anthem to play when joining a channel
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

// InitDB - attempt to make a connection to the Mongo database
func InitDB() {
	// Client parameters
	var err error
	clientOptions := options.Client().ApplyURI(botConfig.DBConfig.URI)
	client, err = mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Println("Error creating MongoDB client")
		os.Exit(1)
	}

	// Connect to MongoDB
	err = client.Connect(context.Background())
	if err != nil {
		fmt.Println("Error connecting to the database")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	db = client.Database(botConfig.DBConfig.DBName)
	k.userCollection = db.Collection("users")
	k.serverCollection = db.Collection("servers")
	k.reminderCollection = db.Collection("reminders")
	fmt.Println("Connected to MongoDB!")
}

//----- G U I L D   M A N A G E M E N T -----

// GetGuild - Get the server information from the db
func (k *KDB) GetGuild(s *discordgo.Session, id string) (server GuildInfo) {

	filter := bson.D{{"gID", id}}
	err := userCollection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return k.CreateGuild(s, id)
	}

	return k.CreateGuild(s, id)
}

// CreateGuild - Create server from given ID
func (k *KDB) CreateGuild(s *discordgo.Session, id string) (server GuildInfo) {

	guild, err := s.Guild(id)
	if err != nil {
		panic(err)
	}

	// Guild not found - Create new
	server.GID = id
	server.Name = guild.Name
	// Set emotes to default
	server.Emotes.UPVOTE = "⬆"
	server.Emotes.DOWNVOTE = "⬇"
	// Set prefix to default
	server.Config.Prefix = "k!"
	// Set votes to default minimum
	server.Config.MinVotes = 3

	// Add the new server
	objID, err := k.GuildColl.InsertOne(context.Background(), server)
	if err != nil {
		panic(err)
	}

	LogTxt(s, "INFO", fmt.Sprintf("Guild \"%s\" [%s] inserted into DB (MongoID#%S)",
		server.Name, server.ID, objID.InsertedID))
}

//----- U S E R   M A N A G E M E N T -----

// GetUser - Query database for user, creating a new one if none exists
func (k *KDB) GetUser(s *discordgo.Session, id string) (user UserInfo) {
	filter := bson.D{{"userID", id}}
	err := userCollection.FindOne(context.Background(), filter).Decode(&user)
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
	user.UserID = id
	user.Name = discordUser.Username
	user.Discriminator = discordUser.Discriminator
	// Set defaults
	user.PlayAnthem = false
	user.DoneDailies = false

	// Insert user into collection
	objID, err := userCollection.InsertOne(context.Background(), user)
	if err != nil {
		panic(err)
	}
	LogTxt(s, "INFO", fmt.Sprintf("User \"%s\" [%s] inserted into DB (MongoID#%S)",
		user.Name, user.UserID, objID.InsertedID))

	return
}

// // Query data
// func (k *KDB) Query() {
// 	cur, err := userCollection.Find(context.Background(), bson.D{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer cur.Close(context.Background())
// 	for cur.Next(context.Background()) {
// 		var result bson.M
// 		err := cur.Decode(&result)
// 		if err != nil {
// 			panic(err)
// 		}
// 		// do something with result....
// 		fmt.Println(cur.Current)
// 	}
// 	if err := cur.Err(); err != nil {
// 		panic(err)
// 	}
// }
