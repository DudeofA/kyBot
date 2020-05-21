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
	"strings"

	"github.com/bwmarrin/discordgo"
)

func runCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, data string) {

	switch command {

	//----- A C C O U N T -----
	// Get amount of coins in players account
	case "account", "acc":
		msgUser := k.kdb.ReadUser(s, m.Author.ID)
		msgGuild := k.kdb.ReadGuild(s, m.GuildID)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("💵 | You have a total of **%d** %scoins", msgUser.Credits, msgGuild.Currency))
		break

	//----- A G E -----
	// Get the age of the (user, channel, guild) ID entered as the argument, or the message creator
	case "age":
		var msg string
		// If no arguments, return age of sender
		if data == "" {
			msg = GetAge(m.Author.Mention())
		} else {
			// Attempt to get age of argument
			msg = GetAge(data)
		}

		s.ChannelMessageSend(m.ChannelID, msg)
		break

	//----- C O M P E N S A T I O N -----
	// Give 200 coins to everyone who has coins
	case "compensation", "comp":
		if CheckAdmin(s, m) {
			CompDailies(s)
			s.ChannelMessageSend(m.ChannelID, "Users compensated")
		}

	//----- C O N F I G -----
	// Modify or reload config
	case "config", "c":
		if CheckAdmin(s, m) {
			if strings.ToLower(data) == "reload" {
				k.botConfig.Update()
				s.ChannelMessageSend(m.ChannelID, "Updated KDB and botConfig")
			} else if strings.HasPrefix(strings.ToLower(data), "edit") {
				//EditConfig(s, m)
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\nPossible Commands:\n* reload\n* edit```"))
			}
		}

	//----- D A I L I E S -----
	// Gets daily Coins
	case "dailies", "day":
		msgUser := k.kdb.ReadUser(s, m.Author.ID)
		msgGuild := k.kdb.ReadGuild(s, m.GuildID)
		msgUser.CollectDailies(s, m, msgGuild)

		break

	//----- D A R L I N G -----
	// Posts best girl gif
	case "darling", "02":
		embedMsg := &discordgo.MessageEmbed{Description: "Zero Tuwu", Color: 0xfa00ff,
			Image: &discordgo.MessageEmbedImage{URL: "https://cdn.discordapp.com/emojis/496406418962776065.gif"}}
		s.ChannelMessageSendEmbed(m.ChannelID, embedMsg)
		break

	//----- G A M B L E -----
	// Gamble away your coins
	case "gamble", "slots":
		Slots(s, m, data)
		break

	//----- H A N G M A N -----
	// Play hangman
	case "hangman", "hm":
		HangmanGame(s, m, data)
		break

	//----- H E L P -----
	// Display the readme file
	case "help", "h":
		readme, err := ioutil.ReadFile(k.state.pwd + "/README.md")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error openning README, contact bot admin for assistance")
			break
		}

		help := strings.SplitAfter(string(readme), "Begin help command here:")
		if len(help) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Misconfigured README, missing `Begin help command here:` to separate commands")
			break
		}

		// Print readme within a code blog to make the formatting work output
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```"+help[1]+"```"))
		break

	//----- I P -----
	// Displayed the external IP of the bot
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
	// Displays the current amount of karma the bot has
	case "karma":
		msgGuild := k.kdb.ReadGuild(s, m.GuildID)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("☯ | Current Karma: %d", msgGuild.Karma))
		break

	//----- K I C K -----
	// Start a vote to kick (disconnect) a user
	case "kick":
		id := strings.TrimPrefix(data, "<@!")
		id = strings.TrimSuffix(id, ">")
		user, err := s.User(id)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, id+" is not a valid user.")
			return
		}
		voteText := fmt.Sprintf("👢 Vote to kick: %s (%s)", user.Username, user.ID)
		StartVote(s, m, voteText, false)

	//----- M I N E C R A F T -----

	// Polls the configured Minecraft Servers to check if they are up and who is playing
	/*
		case "minecraft", "mc":
			UpdateMinecraft(s, m, data)
			break
	*/

	//----- P I N G -----
	// Replies immediately with 'pong' then calculates the difference of the timestamps to get the ping
	case "ping":
		pongMessage, _ := s.ChannelMessageSend(m.ChannelID, "Pong!")
		pingTime, _ := m.Timestamp.Parse()
		pongTime, _ := pongMessage.Timestamp.Parse()
		s.ChannelMessageEdit(m.ChannelID, pongMessage.ID, fmt.Sprintf("Pong! %v", pongTime.Sub(pingTime)))
		break

	//----- Q U O T E -----
	// Begin a vote for a new quote to be added to the list
	case "quote", "q":
		if data != "" {
			StartVote(s, m, data, true)
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Command Syntax: %squote <quote content here>",
				k.botConfig.Prefix))
		}
		break

	//----- Q U O T E L I S T -----
	// List specified quote
	case "quotelist", "ql":
		// Print quote corresponding to the identifier
		quote := k.kdb.ReadQuote(s, m.GuildID, data)
		if data == "" || quote.Identifier == "" {
			QuoteListIDs(s, m.ChannelID)
			return
		}
		quote.Print(s, m.ChannelID)
		break

	//----- Q U O T E R A N D -----
	// Displays a random quote from the database
	case "quoterandom", "qr":
		quote := k.kdb.ReadQuote(s, m.GuildID, "")
		if quote.Identifier == "" {
			s.ChannelMessageSend(m.ChannelID, "No quotes found :(")
			break
		}
		quote.Print(s, m.ChannelID)
		break

	//----- T E S T [ADMIN] -----
	// testing
	case "test":
		if CheckAdmin(s, m) {
			s.ChannelMessageSend(m.ChannelID, "Starting testing...")
			err := s.GuildMemberMove(m.GuildID, "144220178853396480", data)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
			s.ChannelMessageSend(m.ChannelID, "Testing finshed.")
		}
		break

	//----- V E R S I O N -----
	// Gets the current version from the readme file and prints it
	case "version", "v":
		ver := k.state.version
		s.ChannelMessageSend(m.ChannelID, ver)
		break

	//----- V O I C E   S E R V E R -----
	// Changes the voice server in case of server outage
	case "voiceserver", "vs":
		//Get guild data
		msgGuild := k.kdb.ReadGuild(s, m.GuildID)

		var gParam discordgo.GuildParams

		switch data {
		case "us-east", "us-west", "us-central", "us-south":
			gParam.Region = data
			_, err := s.GuildEdit(m.GuildID, gParam)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
			msgGuild.Region = data
			// msgGuild.Update(s)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Voice server changed to: *%s*", data))
			break
		case "":
			region := fmt.Sprintf("The server is currently in region: _*%s*_\nTo change it, use %svoiceserver <server name>\nOptions are: \n```\nus-east, us-central, us-south, us-west\n```",
				msgGuild.Region, k.botConfig.Prefix)
			s.ChannelMessageSend(m.ChannelID, region)
			break

		default:
			s.ChannelMessageSend(m.ChannelID, "Invalid voice server region")
		}

	//----- V O T E -----
	// Starts a vote depending on the numbers of options
	case "vote", "poll":
		StartVote(s, m, data, false)

	//----- W O T D -----
	// Prints the word of the day
	case "wotd", "word":
		s.ChannelMessageSend(m.ChannelID, "```py\n@ canoodle /kəˈno͞odl/\n# verb\nkiss and cuddle amorously.\n'she was caught canoodling with her boyfriend'\n```")

	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown command \"%s\"", command))
	}
}
