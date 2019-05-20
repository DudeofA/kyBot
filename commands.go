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
	"time"

	"github.com/bwmarrin/discordgo"
)

func runCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, data string) {

	switch command {

	//----- H E L P -----
	//Display the readme file
	case "embed":
		embedMsg := &discordgo.MessageEmbed{Image: &discordgo.MessageEmbedImage{URL: "https://cdn.discordapp.com/emojis/496406418962776065.gif"}}
		s.ChannelMessageSendEmbed(m.ChannelID, embedMsg)
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

//createEmbed - Create an embedded message and send it to the relavent channel
func createEmbed(title string, color int, desc string, f1name string, f1value, f2name string, f2value, image string, thumbnail string) *discordgo.MessageEmbed {

	//Create and update embeded status message
	embed := &discordgo.MessageEmbed{
		//Author:      &discordgo.MessageEmbedAuthor{},
		Color:       color, //factorio color
		Description: desc,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   f1name,
				Value:  f1value,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   f2name,
				Value:  f2value,
				Inline: true,
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: image,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: thumbnail,
		},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     title,
	}

	return embed
}
