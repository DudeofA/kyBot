/* 	games.go
_________________________________
Code for games of Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//Slots - gamble away your credits in a slot machine
func Slots(s *discordgo.Session, m *discordgo.MessageCreate, data string) {
	var winMultiplier = 5
	var jackpotMultiplier = 10

	//Gamble item string - Jackbot item MUST be at the end
	var slots = []string{":lemon:", ":cherries:", ":eggplant:", ":peach:", ":moneybag:"}

	//Explain rules
	if data == "" {
		usage := "Slots:\n\tUsage: slots <amount to gamble> (amount must be multiple of 10)"
		payouts := fmt.Sprintf("\n\tPayouts: \n\t\t2 of a kind - Nothing lost\n\t\t3 of a kind - %dx wager\n\t\t3 money bags - %dx wager", winMultiplier, jackpotMultiplier)
		options := fmt.Sprintf("\n\tChances: \n\t\tThere are %d options, each of the 3 slots are fully random", len(slots))

		//Print terms
		s.ChannelMessageSend(m.ChannelID, usage+payouts+options)
		return
	}

	//Check wager is a valid number
	wager, err := strconv.Atoi(data)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Not a valid numerical wager: \"%s\"", data))
		return
	}

	//Check wager is a multiple of 10
	if wager%10 != 0 || wager < 10 {
		s.ChannelMessageSend(m.ChannelID, "Wager must be a positive multiple of 10")
		return
	}

	//Check gambler has enough in their account
	gambler, gamblerIndex := kdb.GetUser(s, m.Author.ID)
	//Save credit balance for later - comparison
	originalCredits := gambler.Credits
	if originalCredits < wager {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You only have %d coins when your wager was %d", gambler.Credits, wager))
		return
	}

	//Take wager from user
	kdb.Users[gamblerIndex].Credits -= wager

	//Roll the slots
	slot1 := rand.Intn(len(slots))
	slot2 := rand.Intn(len(slots))
	slot3 := rand.Intn(len(slots))

	//-- Winnings --
	var winnings int
	var result string

	//Check results
	if slot1 == slot2 && slot1 == slot3 {
		//If all 3 are the same
		if slot1 == len(slots)-1 {
			//Jackpot
			winnings = wager*jackpotMultiplier + wager
			result = "WOW JACKPOT - DING DING DING - YOU JUST WON BIG TIME"
		} else {
			//Normal winnings
			winnings = wager*winMultiplier + wager
			result = "YOU WON - CONGRATS - EZ MONEY"
		}
	} else if slot1 == slot2 || slot1 == slot3 || slot2 == slot3 {
		//If 2 matched
		winnings = wager
		result = "You didn't lose anything...try again?"
	} else {
		//Womp womp
		winnings = 0
		result = "How could this happen to me..."
	}

	//Give winnings and write data back
	kdb.Users[gamblerIndex].Credits += winnings
	kdb.Write()

	//Display the slots
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s %s", slots[slot1], slots[slot2], slots[slot3]))

	//Display balance and result message
	balanceNotice := fmt.Sprintf(":dollar: | You now have a total of **%d** coins", kdb.Users[gamblerIndex].Credits)
	if winnings != 0 && winnings != wager {
		balanceNotice = fmt.Sprintf(":dollar: | Old coins balance: **%d** - You won **%d** coins!\n",
			originalCredits, winnings-wager) + balanceNotice
	}
	s.ChannelMessageSend(m.ChannelID, result+"\n"+balanceNotice)
}

//HangmanGame - ...its hangman, in Discord!
func HangmanGame(s *discordgo.Session, m *discordgo.MessageCreate, data string) {
	var usage = "hangman (start, channel, guess <word/phrase>, quit)\nReact with the letter to guess"
	//var alphabet = []string{"🇦", "🇧", "🇨", "🇩", "🇪", "🇫", "🇬", "🇭", "🇮", "🇯", "🇰", "🇱", "🇲", "🇳", "🇴", "🇵", "🇶", "🇷", "🇸", "🇹", "🇺", "🇻", "🇼", "🇽", "🇾", "🇿"}

	gID := GetGuildByID(m.GuildID)
	hmSession := &kdb.Servers[gID].HM

	//Parse the data passed along with the command
	var command string
	var argument string
	dataArray := strings.SplitN(data, " ", 2)
	if len(dataArray) > 0 {
		command = strings.TrimSpace(dataArray[0])
	}

	if len(dataArray) > 1 {
		argument = strings.TrimSpace(dataArray[1])
	}

	switch strings.ToLower(command) {
	//Usage
	case "":
		s.ChannelMessageSend(m.ChannelID, usage)
		if hmSession.GameState > 0 {
			embed := GenerateLinkEmbed(m.GuildID, hmSession)
			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		}
		break

	//Start a game if not started
	case "start":
		//Check if game isn't already started
		if hmSession.GameState > 0 {
			embed := GenerateLinkEmbed(m.GuildID, hmSession)
			s.ChannelMessageSendEmbed(m.ChannelID, embed)
			return
		}

		//Check if game channel is specified
		if hmSession.Channel == "" {
			s.ChannelMessageSend(m.ChannelID, "Please specify a channel first with 'hm channel <channel>'.")
			return
		}

		//Start the game
		hmSession.GameState = 1

		//Generate word and board
		hmSession.Word = HMGenerator(1)

		for i := 0; i < len(hmSession.Word); i++ {
			hmSession.WordState = append(hmSession.WordState, "_")
		}

		hmGame := HMPrintState(s, hmSession)
		hmSession.Message = hmGame.ID
		break

	//Move game to another channel
	case "channel":
		chanID := strings.TrimPrefix(argument, "<#")
		chanID = strings.TrimSuffix(chanID, ">")
		hmChannel, err := s.Channel(chanID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Please provide a valid channel to move the game to")
			return
		}

		//Get permision of bot
		gamePerm, err := s.State.UserChannelPermissions(self.ID, hmChannel.ID)
		if err != nil {
			panic(err)
		}

		//If bot cannot type here, abort
		if (gamePerm&0x40 != 0x40) || (gamePerm&0x800 != 0x800) {
			s.ChannelMessageSend(m.ChannelID, "Bot cannot send messages/add reactions to this channel")
			return
		}

		kdb.Servers[gID].HM.Channel = hmChannel.ID
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Changed game channel to %s", hmChannel.Mention()))
		break

	//Guess final word/phrase
	case "guess":
		hmSession.GameState++
		HMPrintState(s, hmSession)
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		break

	case "quit":
		if hmSession.GameState == 0 {
			s.ChannelMessageSend(m.ChannelID, "There is no game currently running...")
			return
		}

		//Edit game
		s.ChannelMessageEdit(hmSession.Channel, hmSession.Message, fmt.Sprintf("GAMEOVER\nWord: \"%s\"", hmSession.Word))
		embed := GenerateLinkEmbed(m.GuildID, hmSession)
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

		hmSession.GameState = 0
		hmSession.Message = ""
		hmSession.Word = ""
		hmSession.WordState = nil
		break
	}

	kdb.Write()
}

//HMGenerator - Generate random phrase/word for Hangman
func HMGenerator(num int) (phrase string) {
	//If you request a strange amount of
	if num < 1 || num > 5 {
		return ""
	}

	//Open ENTIRE dictionary
	file, err := os.Open("/usr/share/dict/words")
	if err != nil {
		panic(err)
	}

	//Read file into array
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	allWords := strings.Split(string(bytes), "\n")

	//Generate a random phrase of the specified length
	var phraseArray []string
	for i := 0; i < num; i++ {
		phraseArray = append(phraseArray, allWords[rand.Intn(len(allWords))])
	}

	//Combine and remove 's
	phrase = strings.Join(phraseArray, " ")
	return strings.Replace(phrase, "'s", "", -1)
}

//GenerateLinkEmbed - generate a simple embed to link to the current game of Hangman
func GenerateLinkEmbed(guildID string, hmSession *Hangman) (embed *discordgo.MessageEmbed) {
	link := "https://discordapp.com/channels/"
	messageLink := link + guildID + "/" + hmSession.Channel + "/" + hmSession.Message
	embedLink := fmt.Sprintf("Click [here](%s) to jump to the game", messageLink)
	embed = &discordgo.MessageEmbed{
		Color:       0xB134EB,
		Description: embedLink,
	}
	return embed
}

//HMPrintState - prints the current state of the hangman game
func HMPrintState(s *discordgo.Session, hmSession *Hangman) (hmGame *discordgo.Message) {
	if hmSession.GameState <= 0 {
		return
	}

	stages := []string{
		"\n/---|\n|\n|\n|\n|\n",
		"\n/---|\n|   o\n|\n|\n|\n",
		"\n/---|\n|   o\n|   |\n|\n|\n",
		"\n/---|\n|   o\n|  /|\n|\n|\n",
		"\n/---|\n|   o\n|  /|\\\n|\n|\n",
		"\n/---|\n|   o\n|  /|\\\n|  /\n|\n",
		"\n/---|\n|   o\n|  /|\\\n|  / \\\n|\n",
	}

	//Write out the current word status
	var wordPrint string
	for i := 0; i < len(hmSession.WordState); i++ {
		wordPrint += fmt.Sprintf("%s%s", hmSession.WordState[i], " ")
	}
	wordPrint += "\n"

	gameMessage := []string{
		"YOU GOT IT - Enjoy the coins!",
		"GAME OVER - Try again next time...",
		"Game running...react to guess or use 'hm guess <word>'!\n",
	}

	//Assemble and print the game
	var msgIndex int
	if hmSession.GameState < 0 {
		msgIndex = 0
	} else if hmSession.GameState == 0 {
		msgIndex = 1
	} else {
		msgIndex = 2
	}
	game := "```\n" + gameMessage[msgIndex] + wordPrint + stages[hmSession.GameState-1] + "\n```"

	//If the game is just starting
	if hmSession.Message == "" {
		hmGame, _ = s.ChannelMessageSend(hmSession.Channel, game)
		return
	}

	//Otherwise just edit the existing game
	hmGame, err := s.ChannelMessageEdit(hmSession.Channel, hmSession.Message, game)
	if err != nil {
		s.ChannelMessageSend(hmSession.Channel, "Game message is missing :c")
	}
	return
}
