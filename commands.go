/* 	kylixor.go
_________________________________
Main code for Kylixor Discord Bot
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
