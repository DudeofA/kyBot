/* 	commands.go
_________________________________
Parses commands and executes them for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import "time"

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
