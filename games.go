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
	var link = "https://discordapp.com/channels/"
	var messageLink string

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
		if hmSession.State == 1 {
			messageLink = link + m.GuildID + "/" + hmSession.Channel + "/" + hmSession.Message
			usage += fmt.Sprintf("Game is already in progress [here](%s)", messageLink)
		}
		s.ChannelMessageSend(m.ChannelID, usage)
		break

	//Start a game if not started
	case "start":
		//Check if game isn't already started
		if hmSession.State != 0 {
			messageLink := link + m.GuildID + "/" + hmSession.Channel + "/" + hmSession.Message
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Game is already in progress [here](%s)", messageLink))
			return
		}

		//Check if game channel is specified
		if hmSession.Channel == "" {
			s.ChannelMessageSend(m.ChannelID, "Please specify a channel first with 'hm channel <channel>'.")
			return
		}

		//Start the game
		hmSession.State = 1

		//Generate word and board
		hmSession.Word = HMGenerator(1)

		var wordUnderlines []string
		var i int
		for i = 0; i < len(hmSession.Word); i++ {
			wordUnderlines = append(wordUnderlines, "_")
		}

		var wordPrint string
		for i = 0; i < len(wordUnderlines); i++ {
			wordPrint += fmt.Sprintf("%s%s", wordUnderlines[i], " ")
		}
		hmGame, _ := s.ChannelMessageSend(m.ChannelID, "```\n"+wordPrint+"\n```")
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
		kdb.Write()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Changed game channel to %s", hmChannel.Mention()))
		break

	//Guess final word/phrase
	case "guess":

		break

	case "quit":
		hmSession.State = 0
		s.ChannelMessageEdit(hmSession.Channel, hmSession.Message, fmt.Sprintf("GAMEOVER\nWord: \"%s\"", hmSession.Word))
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("End result [here](%s)", messageLink))
		var embed *discordgo.MessageEmbed
		embed.Description = "[test](https://google.com)"
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		break
	}
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
