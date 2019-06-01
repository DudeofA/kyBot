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
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"
)

func runCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, data string) {

	switch command {

	//----- A C C O U N T -----
	//Get amount of coins in players account
	case "account":
		user, _ := kdb[GetGuildByID(m.GuildID)].GetUserData(s, m)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("💵 | You have a total of **%d** %scoins", user.Credits, config.Coins))
		break

	//----- C O N F I G -----
	//Modify or reload config
	case "config":
		if CheckAdmin(s, m) {
			switch data {
			case "reload":
				UpdateUserFile()
				break

			default:
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\nPossible Commands:\n1. reload\n```"))
			}
		}

	//----- D A I L I E S -----
	//Gets daily Coins
	case "dailies":
		//Retrieve user data from memory
		_, index := kdb[GetGuildByID(m.GuildID)].GetUserData(s, m)
		userData := &kdb[GetGuildByID(m.GuildID)].Users[index]
		//If the dailies have not been done
		if !userData.Dailies {
			//Mark dailies as done and add the appropriate amount
			userData.Dailies = true
			userData.Credits += 100
			//Indicate to user they have recived their dailies
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
				"💵 | Dailies received! Total %scoins: **%d**", config.Coins, userData.Credits))
			//Write data back out to the file
			WriteUserFile()
		} else {
			_, nextRuntime := gocron.NextRun()
			timeUntil := time.Until(nextRuntime)
			hour := timeUntil / time.Hour
			timeUntil -= hour * time.Hour
			min := timeUntil / time.Minute
			timeUntil -= min * time.Minute
			sec := timeUntil / time.Second

			hourStr := "s"
			minStr := "s"
			secStr := "s"
			if hour == 1 {
				hourStr = ""
			}
			if min == 1 {
				minStr = ""
			}
			if sec == 1 {
				secStr = ""
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
				"💵 | You have already collected today's dailies.\nDailies reset in %d hour%s, %d minute%s and %d second%s.",
				hour, hourStr, min, minStr, sec, secStr))
		}
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

	//----- I P -----
	//Displayed the external IP of the bot
	case "ip":
		if CheckAdmin(s, m) {
			resp, err := http.Get("http://myexternalip.com/raw")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			responseData, _ := ioutil.ReadAll(resp.Body)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Bot's current external IP: %s", string(responseData)))
		}
		break

	//----- K A R M A -----
	//Displays the current amount of karma the bot has
	case "karma":
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("☯ | Current Karma: %d", kdb[GetGuildByID(m.GuildID)].Karma))
		break

	//----- P I N G -----
	//Replies immediately with 'pong' then calculates the difference of the timestamps to get the ping
	case "ping":
		pongMessage, _ := s.ChannelMessageSend(m.ChannelID, "Pong!")
		pingTime, _ := m.Timestamp.Parse()
		pongTime, _ := pongMessage.Timestamp.Parse()
		s.ChannelMessageEdit(m.ChannelID, pongMessage.ID, fmt.Sprintf("Pong! %v", pongTime.Sub(pingTime)))
		break

	//----- Q U O T E -----
	//Begin a vote for a new quote to be added to the list
	case "quote":
		if data != "" {
			if startVote(s, m, fmt.Sprintf("1 %s", data)) {
				//addQuote
			} else {
				s.ChannelMessageSend(m.ChannelID, "Vote failed, quote will not be saved")
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "Command Syntax: quote <quote content here>")
		}
		break

	//----- V E R S I O N -----
	//Gets the current version from the readme file and prints it
	case "version":
		ver := GetVersion(s)
		s.ChannelMessageSend(m.ChannelID, ver)
		break

	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown command \"%s\"", command))
	}

}
