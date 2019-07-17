/* 	kdb.go
_________________________________
Manipulating 'database' for guild data for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

//----- K D B   S T R U C T U R E -----

//KDB - "Database"
type KDB struct {
	Servers []ServerStats `json:"servers"` //Servers array
	Users   []UserStats   `json:"users"`   //Users array
}

//----- S E R V E R   S T A T S -----

//ServerStats - Hold all the pertaining information for each server
type ServerStats struct {
	Config Config  `json:"config"`  //Guild specific config
	Emotes Emote   `json:"emotes"`  //String of customizable emotes
	GID    string  `json:"gID"`     //discord guild ID
	HM     Hangman `json:"hangman"` //Holds hangman data
	Karma  int     `json:"karma"`   //bots karma - per server
	Quotes []Quote `json:"quotes"`  //Array of quotes
}

//Config - structure to hold variables specifically for that guild
type Config struct {
	Coins    string `json:"coins"`    //Name of currency that bot uses (i.e. <gold> coins)
	Follow   bool   `json:"follow"`   //Whether or not the bot joins/follows into voice channels for anthems
	MinVotes int    `json:"minVotes"` //Minimum upvotes to pass a vote
	Prefix   string `json:"prefix"`   //Prefix the bot will respond to
}

//Emote - customizable emotes for reactions the bot adds
type Emote struct {
	UPVOTE   string `json:"upvote"`   //Upvote emotes
	DOWNVOTE string `json:"downvote"` //Downvote emotes
}

//Hangman - State of hangman game
type Hangman struct {
	Channel string `json:"channel"` //ChannelID where game is played
	State   int    `json:"state"`   //State of game, 1-7 until you lose
	Word    string `json:"word"`    //Word/phrase for the game
}

//Quote - Data about quotes and quotes themselves
type Quote struct {
	Quote     string    `json:"quote"`     //Actual quoted text
	Timestamp time.Time `json:"timestamp"` //Timestamp when quote was recorded
}

//----- U S E R   S T A T S -----

//UserStats - Hold all pertaining information for each user
type UserStats struct {
	Name        string      `json:"name"`        //Username
	UserID      string      `json:"userID"`      //User ID
	CurrentCID  string      `json:"currentCID"`  //Current channel ID
	LastSeenCID string      `json:"lastSeenCID"` //Last seen channel ID
	PlayAnthem  bool        `json:"playAnthem"`  //True if anthem should play when user joins channel
	Anthem      string      `json:"anthem"`      //Anthem to play when joining a channel
	Credits     int         `json:"credits"`     //Credits gained from dailies
	Dailies     bool        `json:"dailies"`     //True if dailies have been claimed today
	Reminders   []Reminders `json:"reminders"`   //Array of reminders
}

//Reminders - holds reminders for the bot to tell the user about
type Reminders struct {
	UserID     string    `json:"userID"`
	RemindTime time.Time `json:"remindTime"`
	RemindMsg  string    `json:"remindMsg"`
}

//----- M A S T E R   D A T A B A S E
//Kylixor """"Database""""
var kdb KDB

//----- U S E R   F I L E   F U N C T I O N S -----

//InitKDB - Create and initialize user data file
func InitKDB() {
	//Indent so its readable
	userData, err := json.MarshalIndent(kdb, "", "    ")
	if err != nil {
		panic(err)
	}
	//Open file
	jsonFile, err := os.Create(pwd + "/data/kdb.json")
	if err != nil {
		panic(err)
	}
	//Write to file
	_, err = jsonFile.Write(userData)
	if err != nil {
		panic(err)
	}
	//Cleanup
	jsonFile.Close()
}

//ReadKDB - Read in the user file into the structure
func (k *KDB) Read() {
	//Open file
	file, err := os.Open(pwd + "/data/kdb.json")
	if err != nil {
		panic(err)
	}

	//Decode JSON and inject into structure
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&k)
	if err != nil {
		panic(err)
	}

	//Close file
	file.Close()
}

//WriteKDB - Write the file
func (k *KDB) Write() {
	//Marshal data to be readable
	jsonData, err := json.MarshalIndent(k, "", "    ")
	if err != nil {
		panic(err)
	}
	//Open file
	jsonFile, err := os.Create(pwd + "/data/kdb.json")
	if err != nil {
		panic(err)
	}
	//Write to file
	_, err = jsonFile.Write(jsonData)
	if err != nil {
		panic(err)
	}
	//Cleanup
	jsonFile.Close()
}

//Update - Read then write the user jsonFile
func (k *KDB) Update() {
	k.Read()
	k.Write()
}

//----- U S E R   M A N A G E M E N T -----

//CreateUser - create user within the user json file and return it
func (k *KDB) CreateUser(s *discordgo.Session, id string) (userData UserStats, index int) {
	var user UserStats

	//Pull user info from discord
	discordUser, _ := s.User(id)

	//Put user data into user structure
	user.Name = discordUser.Username
	user.Credits = 0
	user.UserID = id
	user.PlayAnthem = false

	//Append new user to the users array
	k.Users = append(k.Users, user)
	//Index will be the last index, or length minus 1
	index = len(k.Users) - 1
	//Write to the file to update it and return the data
	k.Write()
	return user, index
}

//GetUser - Retrieve user data
func (k *KDB) GetUser(s *discordgo.Session, id string) (userData UserStats, index int) {

	//Check if user is in the data file, return them if they are
	for i := range k.Users {
		if k.Users[i].UserID == id {
			return k.Users[i], i
		}
	}

	//return user
	return k.CreateUser(s, id)
}

//TODO

//UpdateUser - Update user data json jsonFile
func (u *ServerStats) UpdateUser(s *discordgo.Session, c interface{}) bool {
	//Return true if update was needed
	return false
}

//----- M I S C   F U N C T I O N S -----

//GetGuildByID - Get the correct ServerStats array from the kdb
func GetGuildByID(id string) (index int) {
	for i, server := range kdb.Servers {
		if server.GID == id {
			return i
		}
	}

	//Guild not found - Create new
	var newServer ServerStats
	newServer.GID = id
	//Set emotes to default
	newServer.Emotes.UPVOTE = "⬆"
	newServer.Emotes.DOWNVOTE = "⬇"
	//Set prefix to default
	newServer.Config.Prefix = "k!"
	//Set votes to default minimum
	newServer.Config.MinVotes = 3

	//Append guild to kdb, write it, and return the index of the guild
	kdb.Servers = append(kdb.Servers, newServer)
	newGuildIndex := len(kdb.Servers) - 1
	kdb.Write()
	return newGuildIndex
}
