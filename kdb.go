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

//ServerStats - Hold all the pertaining information for each server
type ServerStats struct {
	Config Config      `json:"config"` //guild specific config
	Emotes Emotes      `json:"emotes"` //String of customizable emotes
	GID    string      `json:"gID"`    //discord guild ID
	Karma  int         `json:"karma"`  //bots karma - per server
	Users  []UserStats `json:"users"`  //Array of users' information
	Quoted []Quotes    `json:"quotes"` //Array of quotes
}

//Config - structure to hold variables specifically for that guild
type Config struct {
	Coins  string `json:"coins"`  //Name of currency that bot uses (i.e. <gold> coins)
	Follow bool   `json:"follow"` //Whether or not the bot joins/follows into voice channels for anthems
	LogID  string `json:"logID"`  //ID of channel for logging
	Prefix string `json:"prefix"` //Prefix the bot will respond to
}

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

//Emotes - customizable emotes for reactions the bot adds
type Emotes struct {
	UPVOTE   string `json:"upvote"`   //Upvote emotes
	DOWNVOTE string `json:"downvote"` //Downvote emotes
}

//Quotes - Data about quotes and quotes themselves
type Quotes struct {
	Quote     string    `json:"quote"`     //Actual quoted text
	Timestamp time.Time `json:"timestamp"` //Timestamp when quote was recorded
}

//Kylixor """"Database""""
var kdb []ServerStats

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
func ReadKDB() {
	//Open file
	file, err := os.Open(pwd + "/data/kdb.json")
	if err != nil {
		panic(err)
	}

	//Decode JSON and inject into structure
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&kdb)
	if err != nil {
		panic(err)
	}

	//Close file
	file.Close()
}

//WriteKDB - Write the file
func WriteKDB() {
	//Marshal global variable data
	jsonData, err := json.MarshalIndent(kdb, "", "    ")
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

//UpdateKDB - Read then write the user jsonFile
func UpdateKDB() {
	ReadKDB()
	WriteKDB()
}

//----- U S E R   M A N A G E M E N T -----

//CreateUser - create user within the user json file and return it
func (u *ServerStats) CreateUser(s *discordgo.Session, c interface{}) (userData UserStats, index int) {
	var user UserStats

	//Temp - assign interface to MessageEmbed
	m := c.(*discordgo.MessageCreate)

	//Pull user info from discord
	discordUser, _ := s.User(m.Author.ID)

	//Put user data into user structure
	user.Name = discordUser.Username
	user.Credits = 0
	user.UserID = m.Author.ID
	user.PlayAnthem = false

	//Append new user to the users array
	u.Users = append(u.Users, user)
	//Index will be the last index, or length minus 1
	index = len(u.Users) - 1
	//Write to the file to update it and return the data
	WriteKDB()
	return user, index
}

//GetUserData - Retrieve user data
func (u *ServerStats) GetUserData(s *discordgo.Session, c interface{}) (userData UserStats, index int) {

	//Temp - assign interface to message
	m := c.(*discordgo.MessageCreate)

	//Check if user is in the data file, return them if they are
	for i := range u.Users {
		if u.Users[i].UserID == m.Author.ID {
			return u.Users[i], i
		}
	}

	//return user
	return u.CreateUser(s, c)
}

//UpdateUser - Update user data json jsonFile
func (u *ServerStats) UpdateUser(s *discordgo.Session, c interface{}) bool {
	//Return true if update was needed
	return false
}

//----- M I S C   F U N C T I O N S -----

//GetGuildByID - Get the correct ServerStats array from the kdb
func GetGuildByID(id string) (index int) {
	for i, ss := range kdb {
		if ss.GID == id {
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

	//Append guild to kdb, write it, and return the index of the guild
	kdb = append(kdb, newServer)
	newGuildIndex := len(kdb) - 1
	WriteKDB()
	return newGuildIndex
}
