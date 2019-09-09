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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Stages of the hanging
var hmStages = []string{
	"\n/---|\n|\n|\n|\n|\n",
	"\n/---|\n|   o\n|\n|\n|\n",
	"\n/---|\n|   o\n|   |\n|\n|\n",
	"\n/---|\n|   o\n|  /|\n|\n|\n",
	"\n/---|\n|   o\n|  /|\\\n|\n|\n",
	"\n/---|\n|   o\n|  /|\\\n|  /\n|\n",
	"\n/---|\n|   o\n|  /|\\\n|  / \\\n|\n",
	"\n/---|\n|   o\n|  /|\\\n| _/ \\\n|\n",
	"\n/---|\n|   o\n|  /|\\\n| _/ \\_\n|\n",
}
var alphaBlocks = []string{"🇦", "🇧", "🇨", "🇩", "🇪", "🇫", "🇬", "🇭", "🇮", "🇯", "🇰", "🇱", "🇲", "🇳", "🇴", "🇵", "🇶", "🇷", "🇸", "🇹", "🇺", "🇻", "🇼", "🇽", "🇾", "🇿"}

// Slots - gamble away your credits in a slot machine
func Slots(s *discordgo.Session, m *discordgo.MessageCreate, data string) {
	var winMultiplier = 10
	var jackpotMultiplier = 100

	// Gamble item string - Jackbot item MUST be at the end
	var slots = []string{":lemon:", ":cherries:", ":eggplant:", ":peach:", ":strawberry:", ":moneybag:"}

	// Explain rules
	if data == "" {
		usage := "Slots:\n\tUsage: slots <amount to gamble> (amount must be multiple of 10)"
		payouts := fmt.Sprintf(
			"\n\tPayouts: \n\t\t2 of a kind - Nothing lost\n\t\t3 of a kind - %dx wager\n\t\t3 money bags - %dx wager",
			winMultiplier, jackpotMultiplier)
		options := fmt.Sprintf(
			"\n\tChances: \n\t\tThere are %d options, each of the 3 slots are fully random",
			len(slots))

		// Print terms
		s.ChannelMessageSend(m.ChannelID, usage+payouts+options)
		return
	}

	// Check wager is a valid number
	wager, err := strconv.Atoi(data)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Not a valid numerical wager: \"%s\"", data))
		return
	}

	// Check wager is a multiple of 10
	if wager%10 != 0 || wager < 10 {
		s.ChannelMessageSend(m.ChannelID, "Wager must be a positive multiple of 10")
		return
	}

	// Check gambler has enough in their account
	gambler := kdb.ReadUser(s, m.Author.ID)
	// Save credit balance for later - comparison
	originalCredits := gambler.Credits
	if originalCredits < wager {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
			"You only have %d coins when your wager was %d",
			gambler.Credits, wager))
		return
	}

	// Take wager from user
	gambler.Credits -= wager

	// Roll the slots (**RANDOM**)
	slot1 := rand.Intn(len(slots))
	slot2 := rand.Intn(len(slots))
	slot3 := rand.Intn(len(slots))

	// Winnings
	var winnings int
	var result string

	// Check results
	if slot1 == slot2 && slot1 == slot3 {
		// If all 3 are the same
		if slot1 == len(slots)-1 {
			// Jackpot
			winnings = wager*jackpotMultiplier + wager
			result = "WOW JACKPOT - DING DING DING - YOU JUST WON BIG TIME"
		} else {
			// Normal winnings
			winnings = wager*winMultiplier + wager
			result = "YOU WON - CONGRATS - EZ MONEY"
		}
	} else if slot1 == slot2 || slot1 == slot3 || slot2 == slot3 {
		// If 2 matched
		winnings = wager
		result = "You didn't lose anything...try again?"
	} else {
		// Womp womp
		winnings = 0
		result = "How could this happen to me..."
	}

	// Give winnings and write data back
	gambler.Credits += winnings
	kdb.UpdateUser(s, gambler)

	// Display the slots
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s %s", slots[slot1], slots[slot2], slots[slot3]))

	// Display balance and result message
	balanceNotice := fmt.Sprintf(":dollar: | You now have a total of **%d** coins", gambler.Credits)
	if winnings != 0 && winnings != wager {
		balanceNotice = fmt.Sprintf(":dollar: | Old coins balance: **%d** - You won **%d** coins!\n",
			originalCredits, winnings-wager) + balanceNotice
	}
	s.ChannelMessageSend(m.ChannelID, result+"\n"+balanceNotice)
}

// HangmanGame - ...its hangman, in Discord!
func HangmanGame(s *discordgo.Session, m *discordgo.MessageCreate, data string) {
	var usage = "```\n----- HANGMAN -----\nhangman (start, channel, guess <word/phrase>, reprint, quit)\nReact with the letter to guess\n```"

	// Parse the data passed along with the command
	var command string
	var argument string
	dataArray := strings.SplitN(data, " ", 2)
	if len(dataArray) > 0 {
		command = strings.TrimSpace(dataArray[0])
	}
	if len(dataArray) > 1 {
		argument = strings.TrimSpace(dataArray[1])
	}

	hmSession := kdb.ReadHM(s, m.GuildID)

	switch strings.TrimSpace(strings.ToLower(command)) {
	// Usage
	case "":
		s.ChannelMessageSend(m.ChannelID, usage)
		if hmSession.GameState > 0 {
			embed := GenerateHMLinkEmbed(m.GuildID, hmSession, "")
			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		}
		break

	//Start a game if not started
	case "start":
		//Check if game isn't already started
		if hmSession.GameState > 0 {
			embed := GenerateHMLinkEmbed(m.GuildID, hmSession, "Game already started...\n")
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
		hmSession.GenerateWord()
		hmSession.UpdateState(s, m.Author.ID)
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

		hmSession.Channel = hmChannel.ID
		kdb.UpdateHM(s, hmSession)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Changed game channel to %s", hmChannel.Mention()))
		break

	// Can guess a word or letter
	case "guess":
		// Try to delete message to keep game clean
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			LogTxt(s, "INFO", "Bot does not have permission to delete messages, hangman may become messy")
		}

		hmSession.Guess(s, argument)
		hmSession.UpdateState(s, m.Author.ID)
		break

	case "reprint":
		if hmSession.GameState > 0 {
			hmSession.Message = ""
			hmSession.UpdateState(s, m.Author.ID)
		}
		break

	case "quit", "stop":
		if hmSession.GameState == 0 {
			s.ChannelMessageSend(m.ChannelID, "There is no game currently running...")
			return
		}

		// Clean up and end game
		embed := GenerateHMLinkEmbed(m.GuildID, hmSession, "Game ended\n")
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		hmSession.GameState = len(hmStages)
		hmSession.UpdateState(s, m.Author.ID)

		hmSession.ResetGame()
		kdb.UpdateHM(s, hmSession)
		break
	}
}

// GenerateWord - Generate random phrase/word for Hangman
func (hmSession *Hangman) GenerateWord() {
	// Open ENTIRE dictionary in unix
	var err error
	file, err := os.Open(filepath.FromSlash("/usr/share/dict/words"))

	// For fithy Windows users
	if err != nil {
		file, err = os.Open(filepath.FromSlash(pwd + "/data/words.txt"))
		if err != nil {
			fmt.Println("Please tell me your operating system if you see this")
			panic(err)
		}
	}

	// Read file into array
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	allWords := strings.Split(string(bytes), "\n")

	// Generate a random phrase of the specified length
	word := allWords[rand.Intn(len(allWords))]
	// Remove 's
	word = strings.Replace(word, "'s", "", -1)

	hmSession.Word = word

	for i := 0; i < len(hmSession.Word); i++ {
		hmSession.WordState = append(hmSession.WordState, "_")
	}
}

// GenerateHMLinkEmbed - generate a simple embed to link to the current game of Hangman
func GenerateHMLinkEmbed(guildID string, hmSession Hangman, note string) (embed *discordgo.MessageEmbed) {
	link := "https://discordapp.com/channels/"
	messageLink := link + guildID + "/" + hmSession.Channel + "/" + hmSession.Message
	embedLink := fmt.Sprintf("%sClick [here](%s) to jump to the game", note, messageLink)
	embed = &discordgo.MessageEmbed{
		Color:       0xB134EB,
		Description: embedLink,
	}
	return embed
}

//UpdateState - prints the current state of the hangman game
func (hmSession Hangman) UpdateState(s *discordgo.Session, authorID string) {

	//Append guesses into one big string
	guesses := "Guesses: "
	for _, guess := range hmSession.Guessed {
		guesses += guess + ", "
	}
	guesses = strings.TrimSuffix(strings.TrimSpace(guesses), ",")
	guesses += "\n"

	//Assemble and print the game and determine stage to draw
	var stage string
	var gameMessage string
	if hmSession.GameState < 0 {
		hmWinnings := len(hmSession.Word) * 10
		winner := kdb.ReadUser(s, authorID)
		gameMessage = fmt.Sprintf("Guessed correctly by %s\n", winner.Name)
		s.ChannelMessageSend(hmSession.Channel, fmt.Sprintf("YOU GOT IT <@%s> - Enjoy the %d coins!\n",
			winner.ID, hmWinnings))
		winner.Credits += hmWinnings
		kdb.UpdateUser(s, winner)

		// Put the word in the underline
		wordSplice := strings.Split(hmSession.Word, "")
		for i := range hmSession.WordState {
			hmSession.WordState[i] = wordSplice[i]
		}
		stage = hmStages[len(hmSession.Guessed)]
	} else if hmSession.GameState == len(hmStages) {
		gameMessage = fmt.Sprintf("GAME OVER - Try again next time...Word was \"%s\"\n", hmSession.Word)
		stage = hmStages[len(hmStages)-1]
	} else {
		gameMessage = "Game running...react a blue letter to guess or use 'hm guess <word>'!\n"
		stage = hmStages[len(hmSession.Guessed)]
	}

	//Write out the current word status
	var wordPrint string
	for i := 0; i < len(hmSession.WordState); i++ {
		wordPrint += fmt.Sprintf("%s%s", hmSession.WordState[i], " ")
	}
	wordPrint += "\n"

	//Assemble the master string of the game status
	game := "```\n" + gameMessage + guesses + wordPrint + stage + "\n```"

	//If the game is just starting
	if hmSession.Message == "" {
		hmGame, _ := s.ChannelMessageSend(hmSession.Channel, game)
		hmSession.Message = hmGame.ID
	} else {
		//Otherwise just edit the existing
		_, err := s.ChannelMessageEdit(hmSession.Channel, hmSession.Message, game)
		if err != nil {
			s.ChannelMessageSend(hmSession.Channel, "Game message is missing :c, please quit and restart the game")
		}
	}

	//If player just won or lost, reset game
	if hmSession.GameState < 0 || hmSession.GameState == len(hmStages) {
		guildName := "N/A"
		channel, err := s.Channel(hmSession.Channel)
		if err == nil {
			guild, _ := s.Guild(channel.GuildID)
			guildName = guild.Name
		}
		PrintLog(s, "HANGMAN", time.Now(), guildName, hmSession.Channel, authorID, "N/A", "Hangman game reseting...")

		hmSession.ResetGame()
	}

	kdb.UpdateHM(s, hmSession)
}

//ResetGame - resets game stats back to defaults
func (hmSession *Hangman) ResetGame() {
	hmSession.GameState = 0
	hmSession.Guessed = nil
	hmSession.Message = ""
	hmSession.Word = ""
	hmSession.WordState = nil
}

//ReactionGuess - processes letter guesses on hangman using reactions
func ReactionGuess(s *discordgo.Session, r *discordgo.MessageReactionAdd, hmSession *Hangman) {
	var alphabet = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	//Get the letter to guess
	var guess string
	for i := range alphaBlocks {
		if r.Emoji.Name == alphaBlocks[i] {
			guess = alphabet[i]
		}
	}

	//Check that it hasn't been guessed yet
	for _, prevGuess := range hmSession.Guessed {
		if strings.ToLower(guess) == prevGuess {
			return
		}
	}

	//Make the guess and update the board
	hmSession.Guess(s, guess)
	hmSession.UpdateState(s, r.UserID)
}

//Guess - Guess word or letter in the given hangman session
func (hmSession *Hangman) Guess(s *discordgo.Session, guess string) {
	//If no game is running, do nothing and return
	if hmSession.GameState == 0 {
		return
	}

	//Cleanup the guess
	guess = strings.TrimSpace(strings.ToLower(guess))

	//Check if guess is a word, then check if it is a letter, else return
	if len(guess) > 1 {
		//Guess is a word
		if guess == strings.ToLower(hmSession.Word) {
			hmSession.GameState = -1
			return //Won game
		}
	} else if len(guess) == 1 {
		//Guess is a letter
		guessLetterArray := []rune(guess)
		guessLetter := guessLetterArray[0]
		if guessLetter < 'a' || guessLetter > 'z' {
			return
		}

		//Mimic reaction
		guessReaction := alphaBlocks[guessLetter-'a']
		s.MessageReactionAdd(hmSession.Channel, hmSession.Message, guessReaction)

		//Check if the guessed letter is in the word and put those matches in the wordState
		correctGuess := false
		wordSplice := strings.Split(hmSession.Word, "")
		for i := range wordSplice {
			//If guess is correct, check if you won, return if not
			if strings.ToLower(wordSplice[i]) == guess {
				hmSession.WordState[i] = wordSplice[i]
				wordStateString := strings.Join(hmSession.WordState, "")
				//If there are no underlines, game is won
				if !strings.Contains(wordStateString, "_") {
					hmSession.GameState = -1
					return //Won game
				}
				correctGuess = true
			}
		}

		if correctGuess {
			return
		}
	} else {
		return //Invalid guess
	}

	hmSession.GameState++
	hmSession.Guessed = append(hmSession.Guessed, guess)
	return
}
