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
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jasonlvhit/gocron"
)

func runCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, data string) {
	guildIndex := GetGuildByID(m.GuildID)

	switch command {

	//----- A C C O U N T -----
	//Get amount of coins in players account
	case "account", "acc":
		user, _ := kdb.GetUserData(s, m)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("💵 | You have a total of **%d** %scoins", user.Credits, kdb.Servers[guildIndex].Config.Coins))
		break

	//----- A G E -----
	//Get the age of the (user, channel, guild) ID entered as the argument, or the message creator
	case "age":
		var msg string
		if data == "" {
			//If no arguments, get create date of user
			t, err := CreationTime(m.Author.ID)
			if err != nil {
				panic(err)
			}
			msg = fmt.Sprintf("Your account was created on %s", t.Format("Jan 2 3:04:05PM 2006"))
		} else {
			//Else try to take argument as string
			id := data
			id = strings.TrimPrefix(id, "<#")
			id = strings.TrimPrefix(id, "<@")
			id = strings.TrimPrefix(id, "!")
			id = strings.TrimSuffix(id, ">")
			t, err := CreationTime(id)
			if err != nil {
				msg = fmt.Sprintf("Not a valid Discord ID: \"%s\"", data)
			} else {
				msg = fmt.Sprintf("The object was created on %s", t.Format("Jan 2 3:04:05PM 2006"))
			}
		}
		s.ChannelMessageSend(m.ChannelID, msg)
		break

	//----- C O N F I G -----
	//Modify or reload config
	case "config", "c":
		if CheckAdmin(s, m) {
			if strings.ToLower(data) == "reload" {
				kdb.Update()
				botConfig.Update()
				s.ChannelMessageSend(m.ChannelID, "Updated KDB and botConfig")
			} else if strings.HasPrefix(strings.ToLower(data), "edit") {
				//EditConfig(s, m)
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\nPossible Commands:\n* reload\n* edit```"))
			}
		}

	//----- D A I L I E S -----
	//Gets daily Coins
	case "dailies", "day":
		//Retrieve user data from memory
		_, index := kdb.GetUserData(s, m)
		userData := &kdb.Users[index]
		//If the dailies have not been done
		if !userData.Dailies {
			//Mark dailies as done and add the appropriate amount
			userData.Dailies = true
			userData.Credits += botConfig.DailyAmt
			//Indicate to user they have recived their dailies
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
				"💵 | Daily %d coins received! Total %scoins: **%d**", botConfig.DailyAmt, kdb.Servers[guildIndex].Config.Coins, userData.Credits))
			//Write data back out to the file
			kdb.Write()
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
	case "darling", "02":
		embedMsg := &discordgo.MessageEmbed{Description: "Zehro Twu", Color: 0xfa00ff,
			Image: &discordgo.MessageEmbedImage{URL: "https://cdn.discordapp.com/emojis/496406418962776065.gif"}}
		s.ChannelMessageSendEmbed(m.ChannelID, embedMsg)
		break

	//----- H E L P -----
	//Display the readme file
	case "help", "h":
		readme, err := ioutil.ReadFile(pwd + "/README.md")
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
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("☯ | Current Karma: %d", kdb.Servers[guildIndex].Karma))
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
	case "quote", "q":
		if data != "" {
			go func() {
				if startVote(s, m, fmt.Sprintf("0 %s", data)) == 0 {
					QuoteAdd(s, m, data)
				}
			}()
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Command Syntax: %squote <quote content here>", kdb.Servers[guildIndex].Config.Prefix))
		}
		break

		//----- Q U O T E L I S T -----
		//List specified quote
	case "quotelist", "ql":
		i, err := strconv.Atoi(data)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Not a valid number (quotelist <quote index number>)")
		} else {
			//Print quote corresponding to the index number
			QuotePrint(s, m, QuoteGet(m, i-1))
		}
		break

	//----- Q U O T E R A N D -----
	//Displays a random quote from the database
	case "quoterandom", "qr":
		QuotePrint(s, m, QuoteGet(m, -1))
		break

	//----- T E S T -----
	//Runs the commands in a file because I have no idea what I'm doing
	case "test":
		if CheckAdmin(s, m) {
			Test(s, m, command, data)
		}
		break

	//----- V E R S I O N -----
	//Gets the current version from the readme file and prints it
	case "version", "v":
		ver := GetVersion()
		s.ChannelMessageSend(m.ChannelID, ver)
		break

	//----- V O I C E   S E R V E R -----
	//Changes the voice server in case of server outage
	case "voiceserver", "vc":
		//Get guild data
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			panic(err)
		}

		var gParam discordgo.GuildParams

		switch data {
		case "us-east", "us-west", "us-central", "us-south":
			gParam.Region = data
			_, err := s.GuildEdit(m.GuildID, gParam)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
			break
		case "":
			region := fmt.Sprintf("The server is currently in region: %s\nTo change it, use %svoiceserver <server name>\nOptions are: \n```\nus-east, us-central, us-south, us-west\n```", guild.Region, kdb.Servers[guildIndex].Config.Prefix)
			s.ChannelMessageSend(m.ChannelID, region)
			break

		default:
			s.ChannelMessageSend(m.ChannelID, "Invalid voice server region")
		}

	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown command \"%s\"", command))
	}
}
