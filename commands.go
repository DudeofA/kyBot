/* 	commands.go
_________________________________
Parses commands and executes them for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"io/ioutil"

	"github.com/bwmarrin/discordgo"
)

func runCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, data string) {

	switch command {

	//----- A C C O U N T -----
	//Get amount of coins in players account
	case "account":
		user := jcc.GetUserData(s, m)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("💵 | You have a total of **%d** %scoins", user.Credits, config.Coins))
		break

	//----- D A R L I N G -----
	//Posts best girl gif
	case "darling":
		embedMsg := &discordgo.MessageEmbed{Description: "Zehro Twu", Color: 0xfa00ff,
			Image: &discordgo.MessageEmbedImage{URL: "https://cdn.discordapp.com/emojis/496406418962776065.gif"}}
		s.ChannelMessageSendEmbed(m.ChannelID, embedMsg)
		break

	//----- H E L P -----
	//Display the readme file
	case "help":
		readme, err := ioutil.ReadFile("README.md")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error openning README, contact bot admin for assistance")
		}

		//Print readme within a code blog to make the formatting work output
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```"+string(readme)+"```"))
		break

	}

}
