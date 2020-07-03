/* 	games.go
_________________________________
Code for games of Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

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
	gambler := k.kdb.ReadUser(s, m.Author.ID)
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
	gambler.Update()

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

// ----- M I N E C R A F T -----

// UpdateMinecraft - poll configured minecraft servers to status, players, MOTD, and other info
func UpdateMinecraft(s *discordgo.Session, m *discordgo.MessageCreate, command string) {

	type MCResponce struct {
		IP             string `json:"ip"`
		Port           string `json:"port"`
		Online         bool   `json:"online"`
		Version        string `json:"version"`
		MOTD           string `json:"motd"`
		CurrentPlayers string `json:"current_players"`
		MaxPlayers     string `json:"max_players"`
		Error          string `json:"error"`
	}

	var response MCResponce

	apiResponse, err := http.Get("https://api.kylixor.com/mc/" + command)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "ERROR: Unable to make request to API"+err.Error())
		return
	}

	data, _ := ioutil.ReadAll(apiResponse.Body)
	json.Unmarshal(data, &response)

	var status = "- DOWN"
	if response.Online {
		status = "+ UP!"
	}

	if response.Error != "" {
		status += "\n- ERROR: " + response.Error
	}

	msg := fmt.Sprintf("```diff\n----- S T A T U S -----\nAddress: %s\nCurrent Server Version: %s\nThe server is:\n%s\nMOTD: %s\nCurrent players: %s / %s```",
		response.IP, response.Version, status, response.MOTD, response.CurrentPlayers, response.MaxPlayers)
	s.ChannelMessageSend(m.ChannelID, msg)

}
