package status

import (
	"kyBot/kyDB"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

type WordleStat struct {
	MessageID      string `gorm:"primaryKey"`
	UserID         string
	ChannelID      string
	Day            int16 // Wordle day number
	Score          int8  // Score out of 6
	BlankCount     int8  // Black or white squares
	YellowCount    int8
	GreenCount     int8
	FirstWordScore int8 // Blank=0, Yellow=1, Green=2; sum of first row
}

func AddWordleStats(s *discordgo.Session, m *discordgo.Message) (added bool) {
	regex := regexp.MustCompile(`Wordle (\d*) (\w)\/6`)
	if !regex.MatchString(m.Content) {
		log.Debug("Message does not match Wordle regex: \n%s", m.Content)
		return false
	}

	wordleStat := WordleStat{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
		UserID:    m.Author.ID,
	}

	data := regex.FindStringSubmatch(m.Content)

	day, err := strconv.ParseInt(data[1], 10, 16)
	if err != nil {
		log.Errorf("Error converting Wordle day to int: %s", err.Error())
		return false
	}
	wordleStat.Day = int16(day)

	// Make sure stats don't get recorded twice
	var existing *WordleStat
	result := kyDB.DB.Limit(1).Where(WordleStat{UserID: wordleStat.UserID, Day: wordleStat.Day}).Find(&existing)
	if result.RowsAffected >= 1 {
		log.Debugf("User %s already submitted their Wordle for day %d", wordleStat.UserID, wordleStat.Day)
		return false
	}

	score, err := strconv.ParseInt(data[2], 10, 8)
	if err != nil {
		if data[2] == "X" {
			wordleStat.Score = 0
		} else {
			log.Errorf("Error converting Wordle day to int: %s", err.Error())
			return false
		}
	}
	wordleStat.Score = int8(score)

	rows := strings.Split(m.Content, "\n")
	squares := rows[2:]
	for i, row := range squares {
		yellows := int8(strings.Count(row, WORDLE_YELLOW_SQUARE))
		greens := int8(strings.Count(row, WORDLE_GREEN_SQUARE))
		wordleStat.YellowCount += yellows
		wordleStat.GreenCount += greens
		wordleStat.BlankCount += WORDLE_ROW_LENGTH - greens - yellows

		if i == 0 {
			wordleStat.FirstWordScore = WORDLE_GREEN_SCORE*greens + WORDLE_YELLOW_SCORE*yellows
		}
	}

	var wordle Wordle
	result = kyDB.DB.Limit(1).Find(&wordle, Wordle{ChannelID: m.ChannelID})
	if result.RowsAffected != 1 {
		log.Debug("No Wordle game found in this channel")
		return false
	}

	err = s.MessageReactionAdd(m.ChannelID, m.ID, WORDLE_ACK_EMOJI)
	if err != nil {
		log.Errorf("Unable to add reaction to Wordle game results: %s", err.Error())
		return false
	}

	kyDB.DB.Create(&wordleStat)
	return true
}

func (wordle *Wordle) CatchUp(s *discordgo.Session) {
	after := ""
	if len(wordle.Stats) > 0 {
		sort.Slice(wordle.Stats, func(i, j int) bool {
			return wordle.Stats[i].MessageID > wordle.Stats[j].MessageID
		})
		after = wordle.Stats[0].MessageID
	} else {
		after = wordle.StatusMessageID
	}
	messages, err := s.ChannelMessages(wordle.ChannelID, 100, "", after, "")
	if err != nil {
		log.Errorf("Unable to fetch channel messages from %s: %s", wordle.ChannelID, err.Error())
		return
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})
	log.Debugf("Looking through %d missed messages for missed Wordle stats after messageID [%s]", len(messages), after)
	for _, message := range messages {
		if strings.HasPrefix(message.Content, "Wordle") {
			AddWordleStats(s, message)
		}
	}
}

// Go through entire channel and attempt to add previous Wordles
func ScrapeChannel(s *discordgo.Session, m *discordgo.Message) {
	var wordle Wordle
	result := kyDB.DB.Preload(clause.Associations).Find(&wordle, Wordle{ChannelID: m.ChannelID})
	if result.RowsAffected != 1 {
		log.Errorf("Wordle not found in this channel: %s", m.ChannelID)
		return
	}

	// Start looking for messages before the earliest wordle stat
	before := ""
	if len(wordle.Stats) > 0 {
		sort.Slice(wordle.Stats, func(i, j int) bool {
			return wordle.Stats[i].MessageID < wordle.Stats[j].MessageID
		})
		before = wordle.Stats[0].MessageID
	}

	foundWordle := true
	for foundWordle {
		foundWordle = false

		messages, err := s.ChannelMessages(wordle.ChannelID, 100, before, "", "")
		if err != nil {
			log.Errorf("Unable to fetch channel messages from %s: %s", wordle.ChannelID, err.Error())
		}
		if len(messages) == 0 {
			log.Debug("No messages before oldest Wordle Stat")
			return
		}
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].ID < messages[j].ID
		})

		for _, message := range messages {
			if strings.HasPrefix(message.Content, "Wordle") {
				var wordle_stat *WordleStat
				if result := kyDB.DB.Limit(1).Find(&wordle_stat, WordleStat{MessageID: message.ID}); result.RowsAffected == 0 {
					if AddWordleStats(s, message) {
						foundWordle = true
					}
				}
			}
		}
	}
}
