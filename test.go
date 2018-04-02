package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bwmarrin/discordgo"
)

func Test(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Starting testing...")
	//
	//Make sure it actually starts and runs once

	// var jsonBlob = []byte(`{
	// 	"gID": "gid123",
	// 	"user":[{
	// 		"Name": "testname",
	// 		"UserID": "test1234ID",
	// 		"CurrentCID": "channelID123",
	// 		"LastSeenCID": "LSCID",
	// 		"Anthem": "yee",
	// 		"NoiseCredits": 0
	// 	}]
	// }`)

	// USA := UserStateArray{"gid123", "users":["testname", "testID", "testcid", "testlastcid", "yee", 0]}

	var USA UserStateArray
	var user1 = UserState{"testname", "testID", "testcid", "testlastcid", "yee", 0}
	var user2 = UserState{"testname2", "testID2", "testcid2", "testlastcid2", "yee2", 1}
	USA.Users = append(USA.Users, user1)
	USA.Users = append(USA.Users, user2)
	USA.GID = "gid123"

	jsonFinal, err := json.MarshalIndent(USA, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Marshalled JSON: %s\n", jsonFinal)

	//WRITE
	jsonFile, err := os.Create("test.json")
	if err != nil {
		panic(err)
	}
	_, err = jsonFile.Write(jsonFinal)
	if err != nil {
		panic(err)
	}
	jsonFile.Close()

	var USARead UserStateArray
	//READ
	file, err := os.Open("test.json")
	if err != nil {
		panic(err)
	}
	raw, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(raw, &USARead)
	if err != nil {
		panic(err)
	}

	fmt.Printf("From file: %s\n", USARead)
	// fmt.Printf("%d", len(USARead.Users))
	// fmt.Printf("%s", USARead.GID)

	for i := range USARead.Users {
		// s.ChannelMessageSend(config.LogID, fmt.Sprintf("USArray: %s | v.UserID: %s", USArray.Users[i].Name, v.UserID))
		fmt.Printf("Name: %s\n", USARead.Users[i].Name)
	}

	// channel, _ := s.State.Channel(m.ChannelID)
	// guild, _ := s.State.Guild(channel.GuildID)
	// voiceStates := guild.VoiceStates
	// for v := range voiceStates {
	// 	user, _ := s.User(voiceStates[v].UserID)
	// 	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("User#%d: %s", v, user.Username))
	// }

	//Makes sure it makes it to then end
	//
	s.ChannelMessageSend(m.ChannelID, "Testing Complete.")
}
