/* 	kdb.go
_________________________________
Manipulating 'database' for guild data for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bwmarrin/discordgo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//----- M A S T E R   D A T A B A S E
// Kylixor """"Database""""
var kdb KDB
var client *mongo.Client // Database client connection

//----- G L O B A L S -----
var botDatabase = "discord"
var userCollection = "users"
var serverCollection = "servers"

//----- K D B   S T R U C T U R E -----

// KDB - "Database"
type KDB struct {
	Servers []ServerStats `json:"servers" bson:"servers"` // Servers array
	Users   []UserStats   `json:"users" bson:"users"`     // Users array
}

//----- S E R V E R   S T A T S -----

// ServerStats - Hold all the pertaining information for each server
type ServerStats struct {
	Config Config  `json:"config" bson:"config"`   // Guild specific config
	Emotes Emote   `json:"emotes" bson:"emotes"`   // String of customizable emotes
	GID    string  `json:"gID" bson:"gID"`         // discord guild ID
	HM     Hangman `json:"hangman" bson:"hangman"` // Holds hangman data
	Karma  int     `json:"karma" bson:"karma"`     // bots karma - per server
	Quotes []Quote `json:"quotes" bson:"quotes"`   // Array of quotes
}

// Config - structure to hold variables specifically for that guild
type Config struct {
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

// UserStats - Hold all pertaining information for each user
type UserStats struct {
	Name        string      `json:"name" bson:"name"`               // Username
	UserID      string      `json:"userID" bson:"userID"`           // User ID
	CurrentCID  string      `json:"currentCID" bson:"currentCID"`   // Current channel ID
	LastSeenCID string      `json:"lastSeenCID" bson:"lastSeenCID"` // Last seen channel ID
	PlayAnthem  bool        `json:"playAnthem" bson:"playAnthem"`   // True if anthem should play when user joins channel
	Anthem      string      `json:"anthem" bson:"anthem"`           // Anthem to play when joining a channel
	Credits     int         `json:"credits" bson:"credits"`         // Credits gained from dailies
	Dailies     bool        `json:"dailies" bson:"dailies"`         // True if dailies have been claimed today
	Reminders   []Reminders `json:"reminders" bson:"reminders"`     // Array of reminders
}

// Reminders - holds reminders for the bot to tell the user about
type Reminders struct {
	UserID     string    `json:"userID" bson:"userID"`         // User that saved the reminder
	RemindTime time.Time `json:"remindTime" bson:"remindTime"` // Time to remind user
	RemindMsg  string    `json:"remindMsg" bson:"remindMsg"`   // Message to be reminded of
}

//----- U S E R   F I L E   F U N C T I O N S -----

// InitKDB - Create and initialize user data file
func InitKDB() {
	// Indent so its readable
	userData, err := json.MarshalIndent(kdb, "", "    ")
	if err != nil {
		panic(err)
	}
	// Open file
	jsonFile, err := os.Create(filepath.FromSlash(pwd + "/data/kdb.json"))
	if err != nil {
		panic(err)
	}
	// Write to file
	_, err = jsonFile.Write(userData)
	if err != nil {
		panic(err)
	}
	// Cleanup
	jsonFile.Close()
}

// ReadKDB - Read in the user file into the structure
func (k *KDB) Read() {
	// Open file
	file, err := os.Open(filepath.FromSlash(pwd + "/data/kdb.json"))
	if err != nil {
		panic(err)
	}

	// Decode JSON and inject into structure
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&k)
	if err != nil {
		panic(err)
	}

	// Close file
	file.Close()
}

// WriteKDB - Write the file
func (k *KDB) Write() {
	// Marshal data to be readable
	jsonData, err := json.MarshalIndent(k, "", "    ")
	if err != nil {
		panic(err)
	}
	// Open file
	jsonFile, err := os.Create(filepath.FromSlash(pwd + "/data/kdb.json"))
	if err != nil {
		panic(err)
	}
	// Write to file
	_, err = jsonFile.Write(jsonData)
	if err != nil {
		panic(err)
	}
	// Cleanup
	jsonFile.Close()
}

// Update - Read then write the user jsonFile
func (k *KDB) Update() {
	k.Read()
	k.Write()
}

//----- U S E R   M A N A G E M E N T -----

// CreateUser - create user within the user json file and return it
func (k *KDB) CreateUser(s *discordgo.Session, id string) (userData *UserStats) {
	var user UserStats

	// Pull user info from discord
	discordUser, _ := s.User(id)

	// Put user data into user structure
	user.Name = discordUser.Username
	user.Credits = 0
	user.UserID = id
	user.PlayAnthem = false

	// Append new user to the users array
	k.Users = append(k.Users, user)
	// Write to the file to update it and return the data
	k.Write()
	return &user
}

// GetUser - Retrieve user data
func (k *KDB) GetUser(s *discordgo.Session, id string) (userData *UserStats) {

	// Check if user is in the data file, return them if they are
	for i := range k.Users {
		if k.Users[i].UserID == id {
			return &k.Users[i]
		}
	}

	// return user
	return k.CreateUser(s, id)
}

//UpdateUser - Update user data json jsonFile
func (u *ServerStats) UpdateUser(s *discordgo.Session, c interface{}) bool {
	//Return true if update was needed
	return false
}

//----- M I S C   F U N C T I O N S -----

// GetGuildByID - Get the correct ServerStats array from the kdb
func GetGuildByID(id string) (index int) {
	for i, server := range kdb.Servers {
		if server.GID == id {
			return i
		}
	}

	// Guild not found - Create new
	var newServer ServerStats
	newServer.GID = id
	// Set emotes to default
	newServer.Emotes.UPVOTE = "⬆"
	newServer.Emotes.DOWNVOTE = "⬇"
	// Set prefix to default
	newServer.Config.Prefix = "k!"
	// Set votes to default minimum
	newServer.Config.MinVotes = 3

	// Append guild to kdb, write it, and return the index of the guild
	kdb.Servers = append(kdb.Servers, newServer)
	newGuildIndex := len(kdb.Servers) - 1
	kdb.Write()
	return newGuildIndex
}

//----- M O N G O D B   F U N C T I O N S -----

// InitDB - attempt to make a connection to the Mongo database
func InitDB() {
	// Client parameters
	var err error
	clientOptions := options.Client().ApplyURI(botConfig.DBURI)
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

	fmt.Println("Connected to MongoDB!")
}

// Query data
func (k *KDB) Query() {
	collection := client.Database("discord").Collection("numbers")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		panic(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			panic(err)
		}
		// do something with result....
		fmt.Println(cur.Current)
	}
	if err := cur.Err(); err != nil {
		panic(err)
	}
}

// Insert - make a new addition to the db
func (k *KDB) Insert() {
	// Insert something
	collection := client.Database("discord").Collection("numbers")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	if err != nil {
		panic(err)
	}
	id := res.InsertedID
	fmt.Println(id)
}

// AddUser - insert new user into database in collection 'users'
func (k *KDB) AddUser(user UserStats) {
	// Access correct collection
	collection := client.Database("discord").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert user into collection
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		panic(err)
	}
}

// QueryUser - search for user in database in collection 'users'
func (k *KDB) QueryUser(field string, query string) (user UserStats) {
	// Access correct collection
	collection := client.Database("discord").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//Search for user by the specified field
	filter := bson.D{{field, query}}
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		panic(err)
	}

	return
}
