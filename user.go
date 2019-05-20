/* 	commands.go
_________________________________
Parses commands and executes them for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"encoding/json"
	"os"
	"time"
)

//ServerStats - Hold all the pertaining information for each server
type ServerStats struct {
	GID   string      `json:"gID"`   //discord guild ID
	Karma int         `json:"karma"` //bots karma - per server
	Users []UserStats `json:"users"` //Array of users' information
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

var jcc ServerStats

//InitUserFile - Create and initialize user data file
func InitUserFile() {
	//Indent so its readable
	userData, err := json.MarshalIndent(jcc, "", "    ")
	if err != nil {
		panic(err)
	}
	//Open file
	jsonFile, err := os.Create(pwd + "/data/users.json")
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

//ReadUserFile - Read in the user file into the structure
func (u *ServerStats) ReadUserFile() {
	//Open file
	file, err := os.Open(pwd + "/data/users.json")
	if err != nil {
		panic(err)
	}

	//Decode JSON and inject into structure
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&u)
	if err != nil {
		panic(err)
	}

	//Close file
	file.Close()
}

//WriteUserFile - Write the file
func (u *ServerStats) WriteUserFile() {
	//Marshal global variable data
	jsonData, err := json.MarshalIndent(u, "", "    ")
	if err != nil {
		panic(err)
	}
	//Open file
	jsonFile, err := os.Create(pwd + "/data/users.json")
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

//UpdateUserFile - Read then write the user jsonFile
func (u *ServerStats) UpdateUserFile() {
	u.ReadUserFile()
	u.WriteUserFile()
}
