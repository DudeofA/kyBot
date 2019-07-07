/* 	games.go
_________________________________
Code for games of Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

//Slots - gamble away your credits in a slot machine
func Slots(s *discordgo.Session, m *discordgo.MessageCreate, data string) {
	var winMultiplier = 5
	var jackpotMultiplier = 10

	//Gamble item string - Jackbot item MUST be at the end
	var slots = []string{":green_apple:", ":lemon:", ":cherries:", ":eggplant:", ":peach:", ":moneybag:"}

	//Explain rules
	if data == "" {
		usage := "Slots:\n\tUsage: slots <amount to gamble> (amount must be multiple of 10)"
		payouts := fmt.Sprintf("\n\tPayouts: \n\t\t2 of a kind - Nothing lost\n\t\t3 of a kind - %dx wager\n\t\t3 money bags - %dx wager", winMultiplier, jackpotMultiplier)
		options := fmt.Sprintf("\n\tChances: \n\t\tThere are %d options, and each slot is a simple random function (verify on Github)", len(slots))

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

	//Display the slots
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s %s", slots[slot1], slots[slot2], slots[slot3]))

	//-- Winnings --
	var winnings int
	var result string

	//Check results
	if slot1 == slot2 && slot1 == slot3 {
		//If all 3 are the same
		if slot1 == len(slots)-1 {
			//Jackpot
			winnings = wager * 10
			result = "WOW JACKPOT - DING DING DING - YOU JUST WON BIG TIME"
		} else {
			//Normal winnings
			winnings = wager * 3
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

	//Display balance
	s.ChannelMessageSend(m.ChannelID, result)
	balanceNotice := fmt.Sprintf("Current coins balance: %d", kdb.Users[gamblerIndex].Credits)
	if winnings != 0 && winnings != wager {
		balanceNotice = fmt.Sprintf("Old coins balance: %d - ", originalCredits) + balanceNotice
	}
	s.ChannelMessageSend(m.ChannelID, balanceNotice)
}
