package main

import (
	"encoding/json"
	// "fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

type UserStateArray struct {
	GID   string      `json:"gID"`
	Users []UserState `json:"users"`
}

type UserState struct {
	Name         string `json:"name"`
	UserID       string `json:"userID"`
	CurrentCID   string `json:"currentCID"`
	LastSeenCID  string `json:"lastSeenCID"`
	Anthem       string `json:"anthem"`
	NoiseCredits int    `json:"noiseCredits"`
}

var USArray UserStateArray

func ReadUserFile() {
	file, err := os.Open("users.json")
	if err != nil {
		WriteUserFile()
        file, err = os.Open("users.json")
        if err != nil {
            panic(err)
        }
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&USArray)
	if err != nil {
		panic(err)
	}

	file.Close()
}

func ReadUser(s *discordgo.Session, v *discordgo.VoiceStateUpdate) (UVS UserState, i int) {
	//Search through user array for specific user and return them
	for i := range USArray.Users {
		if USArray.Users[i].UserID == v.UserID {
			return USArray.Users[i], i
		}
	}
	//Or create a new one if they cannot be found
	s.ChannelMessageSend(config.LogID, "Cannot find user...Creating new...")
	return CreateUser(s, v, "VOICE"), len(USArray.Users)
}

func CreateUser(s *discordgo.Session, i interface{}, code string) (UVS UserState) {
	var user UserState

	switch code {

	case "VOICE":
		v := i.(*discordgo.VoiceStateUpdate)
		//Create user
		usr, _ := s.User(v.UserID)
		member, err := s.GuildMember(v.GuildID, v.UserID)
		user.Name = FormatAuthor(usr, member, err)
		user.UserID = v.UserID
		user.CurrentCID = v.ChannelID
		user.LastSeenCID = v.ChannelID
		user.Anthem = ""
		user.NoiseCredits = 1
		break

	default:

	}

	USArray.Users = append(USArray.Users, user)
	WriteUserFile()
	return user
}

func UpdateUser(s *discordgo.Session, i interface{}, code string) bool {
	switch code {

	case "VOICE":
		v := i.(*discordgo.VoiceStateUpdate)
		//Get user object
		user, j := ReadUser(s, v)
		//Update user object
		//If the update is only a change in voice (mute, deafen, etc)
		if user.CurrentCID == v.ChannelID && user.LastSeenCID == v.ChannelID {
			return false
		}
		USArray.Users[j].CurrentCID = v.ChannelID
		if v.ChannelID != "" {
			USArray.Users[j].LastSeenCID = v.ChannelID
		}
		break
	case "MSG":
		//m := i.(*discordgo.MessageCreate)
		//user := ReadUser(s, m)
		break
	default:
		panic("Incorrect code sent to WriteUser")
	}

	WriteUserFile()
	return true
}

func WriteUserFile() {
	//Marshal global variable data
    if USArray == nil {
        USArray
    }
	jsonData, err := json.MarshalIndent(USArray, "", "    ")
	if err != nil {
		panic(err)
	}
	//Open file
	jsonFile, err := os.Create("users.json")
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
