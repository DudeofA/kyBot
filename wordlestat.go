package main

import (
	"errors"
	"fmt"
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
	User           User
	ChannelID      string
	Day            int16 // Wordle day number
	Score          int8  // Score out of 6
	BlankCount     int8  // Black or white squares
	YellowCount    int8
	GreenCount     int8
	FirstWordScore int8 // Blank=0, Yellow=1, Green=2; sum of first row
}

func AddWordleStats(m *discordgo.Message, bypassChanID string) (err error) {
	regex := regexp.MustCompile(`Wordle (\d*) (\w)\/6`)
	if !regex.MatchString(m.Content) {
		err := fmt.Errorf("message does not match wordle regex: \n%s", strings.ToLower(m.Content))
		log.Debug(err)
		return err
	}

	var wordleChannelID string
	if bypassChanID != "" {
		wordleChannelID = bypassChanID
	} else {
		wordleChannelID = m.ChannelID
	}

	wordleStat := WordleStat{
		MessageID: m.ID,
		ChannelID: wordleChannelID,
		UserID:    m.Author.ID,
	}

	data := regex.FindStringSubmatch(m.Content)

	day, err := strconv.ParseInt(data[1], 10, 16)
	if err != nil {
		err := fmt.Errorf("error converting wordle day to int: %s", err.Error())
		log.Error(err)
		return err
	}
	wordleStat.Day = int16(day)

	// Make sure stats don't get recorded twice by checking day
	var existing *WordleStat
	result := db.Limit(1).Where(WordleStat{UserID: wordleStat.UserID, Day: wordleStat.Day}).Find(&existing)
	if result.RowsAffected >= 1 {
		err := fmt.Errorf("User %s already submitted their Wordle for day %d", wordleStat.UserID, wordleStat.Day)
		log.Debug(err)
		return err
	}

	score, err := strconv.ParseInt(data[2], 10, 8)
	if err != nil {
		// If they failed, assign them the Wordle failed score
		if data[2] == "X" {
			score = WORDLE_FAIL_SCORE
		} else {
			err := fmt.Errorf("error converting wordle day to int: %s", err.Error())
			log.Error(err)
			return err
		}
	}
	wordleStat.Score = int8(score)

	// Skip the Wordle score line and the newline to find the actual squares and begin counting
	rows := strings.Split(m.Content, "\n")
	rows_of_squares := rows[2:]
	for i, row := range rows_of_squares {
		yellows := int8(strings.Count(row, WORDLE_YELLOW_SQUARE))
		greens := int8(strings.Count(row, WORDLE_GREEN_SQUARE))
		wordleStat.YellowCount += yellows
		wordleStat.GreenCount += greens
		wordleStat.BlankCount += WORDLE_ROW_LENGTH - greens - yellows

		if i == 0 {
			wordleStat.FirstWordScore = WORDLE_GREEN_SCORE*greens + WORDLE_YELLOW_SCORE*yellows
		}
	}

	if int(wordleStat.BlankCount) == len(rows_of_squares)*WORDLE_ROW_LENGTH {
		log.Errorf("No green or yellows recorded, this game is invalid")
	}

	wordle, err := GetWordle(wordleChannelID)
	if err != nil {
		err := errors.New("no wordle game found in this channel")
		log.Debug(err)
		return err
	}

	err = s.MessageReactionAdd(m.ChannelID, m.ID, WORDLE_ACK_EMOJI)
	if err != nil {
		err := fmt.Errorf("unable to add reaction to wordle game results on messageid: %s\n%s", m.ID, err.Error())
		log.Error(err)
		return err
	}

	result = db.Limit(1).Where(WordleStat{UserID: wordleStat.UserID}).Find(&existing)
	if result.RowsAffected == 0 {
		// User has never played
		user := GetUser(m.Author)
		wordle.Players = append(wordle.Players, &user)
	}

	db.Create(&wordleStat)
	wordle.UpdateStatus()
	return nil
}

func (wordle *Wordle) CatchUp() {
	after := ""
	var stats []WordleStat
	db.Find(&stats, WordleStat{ChannelID: wordle.ChannelID})
	if len(stats) > 0 {
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].MessageID > stats[j].MessageID
		})
		after = stats[0].MessageID
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
			AddWordleStats(message, "")
		}
	}
}

// Go through entire channel and attempt to add previous Wordles
func ScrapeChannel(channelID string) {
	var wordle Wordle
	result := db.Preload(clause.Associations).Find(&wordle, Wordle{ChannelID: channelID})
	if result.RowsAffected != 1 {
		log.Errorf("Wordle not found in this channel: %s", channelID)
		return
	}

	// Start looking for messages before the earliest wordle stat
	before := ""
	var stats []WordleStat
	db.Find(&stats, WordleStat{ChannelID: wordle.ChannelID})
	if len(stats) > 0 {
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].MessageID < stats[j].MessageID
		})
		before = stats[0].MessageID
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
				if result := db.Limit(1).Find(&wordle_stat, WordleStat{MessageID: message.ID}); result.RowsAffected == 0 {
					err := AddWordleStats(message, "")
					if err != nil {
						log.Debugf("Wordle stat not successfully added: %s", err)
					} else {
						foundWordle = true
					}
				}
			}
		}
	}
}

func ImportWordleStat(wordleChannelID string, channelID string, messageID string) (err error) {
	if wordleChannelID == "" || messageID == "" {
		err := errors.New("WordleChannelID or MessageID blank")
		log.Error(err.Error())
		return err
	}
	msg, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		err := fmt.Errorf("error finding message %s in discord: %s", messageID, err.Error())
		log.Error(err)
		return err
	}
	err = AddWordleStats(msg, wordleChannelID)
	if err != nil {
		return err
	}
	return nil
}
