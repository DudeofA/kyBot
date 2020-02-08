/* 	hangman.go
_________________________________
Code for hangman of Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Hangman - State of hangman game
type Hangman struct {
	GuildID   string   `json:"guildID"`   // Guild game is attached to
	ChannelID string   `json:"channel"`   // ChannelID where game is played
	GameState int      `json:"gameState"` // State of game, 1-7 until you lose
	Guessed   []string `json:"guessed"`   // Characters/words that have been guessed
	MessageID string   `json:"message"`   // MessageID of current hangman game
	Word      string   `json:"word"`      // Word/phrase for the game
	WordState []string `json:"hmState"`   // State of game's word
}

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
var alphabet = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

//----- H A N G M A N   M A N A G E M E N T -----

// CreateHM - insert the hangman game into the hangman collection
func (kdb *KDB) CreateHM(s *discordgo.Session, guildID string) (hm Hangman) {
	guessed := strings.Join(hm.Guessed, ",")
	wordState := strings.Join(hm.WordState, "")
	hm.GuildID = guildID

	// Insert game into database
	_, err := k.db.Exec("INSERT INTO hangman (guildID, channelID, messageID, word, gameState, wordState, guessed) VALUES(?,?,?,?,?,?,?)",
		hm.GuildID, hm.ChannelID, hm.MessageID, hm.Word, hm.GameState, wordState, guessed)
	if err != nil {
		panic(err)
	}

	hm.ResetGame()
	LogDB("Hangman Game", hm.Word, hm.GuildID, "created")

	return hm
}

// ReadHM - get hangman session from hangman collection
func (kdb *KDB) ReadHM(s *discordgo.Session, guildID string) (hm Hangman) {
	var guessed string
	var wordState string
	// Search by discord guild ID
	row := k.db.QueryRow("SELECT guildID, channelID, messageID, word, gameState, wordState, guessed FROM hangman WHERE guildID=(?)", guildID)
	err := row.Scan(&hm.GuildID, &hm.ChannelID, &hm.MessageID, &hm.Word, &hm.GameState, &wordState, &guessed)
	switch err {
	case sql.ErrNoRows:
		k.Log("KDB", "Hangman game not found for this guild")
		return k.kdb.CreateHM(s, guildID)
	case nil:
		hm.Guessed = nil
		if guessed != "" {
			hm.Guessed = strings.Split(guessed, ",")
		}
		hm.WordState = strings.Split(wordState, "")
		LogDB("Hangman", hm.Word, hm.GuildID, "read")
		return hm
	default:
		panic(err)
	}
}

// Update [HM] - update hangman game in the database
func (hm *Hangman) Update() {
	guessed := strings.Join(hm.Guessed, ",")
	wordState := strings.Join(hm.WordState, "")
	LogDB("Hangman", hm.Word, hm.GuildID, "updated")

	// Update the guild to the database
	_, err := k.db.Exec("INSERT INTO hangman (guildID, channelID, messageID, word, gameState, wordState, guessed) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE channelID = ?, messageID = ?, word = ?, gameState = ?, wordState = ?, guessed = ?",
		hm.GuildID, hm.ChannelID, hm.MessageID, hm.Word, hm.GameState, wordState, guessed, hm.ChannelID, hm.MessageID, hm.Word, hm.GameState, wordState, guessed)
	if err != nil {
		k.Log("FATAL", "Error updating hangman game: "+hm.GuildID)
		panic(err)
	}
}

//----- H A N G M A N   G A M E -----

// HangmanGame - ...its hangman, in Discord!
// Takes in the command an executes it according to the arguments
func HangmanGame(s *discordgo.Session, m *discordgo.MessageCreate, data string) {

	// Parse the data passed along with the command
	var command string
	var argument string
	dataArray := strings.SplitN(data, " ", 2)
	// Get command if there is one (start, guess, quit, etc.)
	if len(dataArray) > 0 {
		command = strings.TrimSpace(dataArray[0])
	}
	// Get argument to command if exists (guess _a_, channel _#489301_)
	if len(dataArray) > 1 {
		argument = strings.TrimSpace(dataArray[1])
	}

	// Read the hangman game from the db, creating new if necessary
	hm := k.kdb.ReadHM(s, m.GuildID)

	// Determine which command has been called
	switch strings.TrimSpace(strings.ToLower(command)) {
	case "":
		var usage = "```\n----- HANGMAN -----\nhangman (channel, guess <word/letter>, quit)\nReact with the letter to guess\n```"

		// If the game is running, reprint the game
		if hm.GameState <= 0 {
			// Check if game channel is specified, then start game
			if hm.ChannelID == "" {
				s.ChannelMessageSend(m.ChannelID, "```\nNo channel set, choosing current channel.  Change with !hm channel #channel\n```")
				hm.ChannelID = m.ChannelID
			}

			s.ChannelMessageSend(hm.ChannelID, usage)

			// Start the game
			hm.GameState = 1

			//Generate word and board
			hm.GenerateWord()
			hm.PrintState(s)
		} else {
			hm.MessageID = ""
			hm.PrintState(s)
		}
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
		gamePerm, err := s.State.UserChannelPermissions(k.state.self.ID, hmChannel.ID)
		if err != nil {
			panic(err)
		}

		//If bot cannot type here, abort
		if (gamePerm&0x40 != 0x40) || (gamePerm&0x800 != 0x800) {
			s.ChannelMessageSend(m.ChannelID, "Bot cannot send messages/add reactions to this channel")
			return
		}

		hm.ChannelID = hmChannel.ID
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Changed game channel to %s", hmChannel.Mention()))
		break

	// Can guess a word or letter
	case "guess", "g":
		// Try to delete message to keep game clean
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			k.Log("INFO", "Bot does not have permission to delete messages, hangman may become messy")
		}

		hm.Guess(s, m.Author.ID, argument)
		break

	case "quit", "stop":
		if hm.GameState == 0 {
			s.ChannelMessageSend(m.ChannelID, "There is no game currently running...")
			return
		}

		// Clean up and end game
		hm.GameState = len(hmStages)
		hm.PrintState(s)
		hm.ResetGame()
		hm.Update()
		break
	}
}

//Guess - Guess word or letter in the given hangman session
func (hm *Hangman) Guess(s *discordgo.Session, authorID string, guess string) {
	//If no game is running, do nothing and return
	if hm.GameState == 0 {
		return
	}

	//Cleanup the guess
	guess = strings.TrimSpace(strings.ToLower(guess))

	//Check if guess is a word, then check if it is a letter, else return
	if len(guess) > 1 {
		//Guess is a word
		if guess == strings.ToLower(hm.Word) {
			hm.WinGame(s, authorID)
			return
		}
	} else if len(guess) == 1 {
		// Check guess is a letter
		guessLetterArray := []rune(guess)
		guessLetter := guessLetterArray[0]
		if guessLetter < 'a' || guessLetter > 'z' {
			return
		}

		// Mimic reaction
		guessReaction := alphaBlocks[guessLetter-'a']
		s.MessageReactionAdd(hm.ChannelID, hm.MessageID, guessReaction)

		lowerWord := strings.ToLower(hm.Word)
		if strings.Index(lowerWord, guess) != -1 {
			// Check if the guessed letter is in the word and put those matches in the wordState
			for {
				index := strings.Index(lowerWord, guess)
				if index == -1 {
					break
				}

				hm.WordState[index] = guess
				lowerWord = lowerWord[:index] + "#" + lowerWord[index+1:]
			}

			//If there are no underlines, game is won
			wordState := strings.Join(hm.WordState, "")
			if !strings.Contains(wordState, "_") {
				hm.WinGame(s, authorID)
				return
			}

			// The letter was correct, return without appending to guesses
			hm.PrintState(s)
			hm.CheckState(s)
			return
		}

	} else {
		return //Invalid guess
	}

	hm.GameState++
	hm.Guessed = append(hm.Guessed, guess)
	hm.PrintState(s)
	hm.CheckState(s)
}

//ReactionGuess - processes letter guesses on hangman using reactions
func (hm *Hangman) ReactionGuess(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	//Get the letter to guess
	var guess string
	for i := range alphaBlocks {
		if r.Emoji.Name == alphaBlocks[i] {
			guess = alphabet[i]
		}
	}

	//Check that it hasn't been guessed yet
	for _, prevGuess := range hm.Guessed {
		if strings.ToLower(guess) == prevGuess {
			return
		}
	}

	//Make the guess and update the board
	hm.Guess(s, r.UserID, guess)
}

//PrintState - prints the current state of the hangman game
func (hm Hangman) PrintState(s *discordgo.Session) {

	//Append guesses into one big string
	guesses := "Wrong guesses:"
	if len(hm.Guessed) > 0 {
		guesses = "Wrong guesses: " + strings.Join(hm.Guessed, ", ")
	}
	guesses += "\n"

	//Assemble and print the game and determine stage to draw
	var stage string
	var gameMessage string
	if hm.GameState < 0 {
		gameMessage = fmt.Sprintf("Guessed correctly\n")

		// Put the word in the underline
		hm.WordState = strings.Split(hm.Word, "")

		stage = hmStages[len(hm.Guessed)]
	} else if hm.GameState == len(hmStages) {
		gameMessage = fmt.Sprintf("GAME OVER - Try again next time...Word was \"%s\"\n", hm.Word)
		stage = hmStages[len(hmStages)-1]
	} else {
		gameMessage = "Game running...react a blue letter to guess or use 'hm guess <word>'\n"
		stage = hmStages[len(hm.Guessed)]
	}

	//Write out the current word status
	wordPrint := strings.Join(hm.WordState, " ") + "\n"

	//Assemble the master string of the game status
	game := "```\n" + gameMessage + guesses + wordPrint + stage + "\n```"

	//If the game is just starting
	if hm.MessageID == "" {
		hmGame, _ := s.ChannelMessageSend(hm.ChannelID, game)
		hm.MessageID = hmGame.ID
	} else {
		//Otherwise just edit the existing
		_, err := s.ChannelMessageEdit(hm.ChannelID, hm.MessageID, game)
		if err != nil {
			s.ChannelMessageSend(hm.ChannelID, "Game message is missing :c, please quit and restart the game")
		}
	}

	hm.Update()
}

// GenerateWord - Generate random phrase/word for Hangman
func (hm *Hangman) GenerateWord() {
	// Open ENTIRE dictionary in unix
	var err error
	file, err := os.Open(filepath.FromSlash("/usr/share/dict/words"))

	// For fithy Windows users
	if err != nil {
		file, err = os.Open(filepath.FromSlash(k.state.pwd + "/dict/words.txt"))
		if err != nil {
			fmt.Println("Cannot find words dictionary or words.txt file, aborting")
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

	hm.Word = word

	for i := 0; i < len(hm.Word); i++ {
		hm.WordState = append(hm.WordState, "_")
	}
}

// GenerateHMLinkEmbed - generate a simple embed to link to the current game of Hangman
func GenerateHMLinkEmbed(guildID string, hm Hangman, note string) (embed *discordgo.MessageEmbed) {
	link := "https://discordapp.com/channels/"
	messageLink := link + guildID + "/" + hm.ChannelID + "/" + hm.MessageID
	embedLink := fmt.Sprintf("%sClick [here](%s) to jump to the game", note, messageLink)
	embed = &discordgo.MessageEmbed{
		Color:       0xB134EB,
		Description: embedLink,
	}
	return embed
}

//ResetGame - resets game stats back to defaults
func (hm *Hangman) ResetGame() {
	hm.GameState = 0
	hm.Guessed = nil
	hm.MessageID = ""
	hm.Word = ""
	hm.WordState = nil
	hm.Update()
}

// WinGame - win the game and award the credits to the user
func (hm *Hangman) WinGame(s *discordgo.Session, authorID string) {
	hm.GameState = -1
	hmWinnings := len(hm.Word) * 10
	winner := k.kdb.ReadUser(s, authorID)
	s.ChannelMessageSend(hm.ChannelID, fmt.Sprintf("YOU GOT IT <@%s> - Enjoy the %d coins!\n",
		winner.ID, hmWinnings))
	winner.Credits += hmWinnings
	winner.Update()

	// Print the final result of the game
	hm.PrintState(s)
	hm.ResetGame()
}

// CheckState - check for a won or lost game
func (hm *Hangman) CheckState(s *discordgo.Session) {
	if hm.GameState < 0 || hm.GameState == len(hmStages) {
		guildName := "N/A"
		channel, err := s.Channel(hm.ChannelID)
		if err == nil {
			guild, _ := s.Guild(channel.GuildID)
			guildName = guild.Name
		}
		k.Log("HANGMAN", fmt.Sprintf("%s - %s - %s - %s", guildName, hm.ChannelID, "", "Hangman game reseting..."))

		hm.ResetGame()
	}
}
