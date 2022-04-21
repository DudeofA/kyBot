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
)

type WordleStat struct {
	MessageID      string `gorm:"primaryKey"`
	UserID         string
	ChannelID      string
	Day            uint16 // Wordle day number
	Score          uint8  // Score out of 6
	BlankCount     uint8  // Black or white squares
	YellowCount    uint8
	GreenCount     uint8
	FirstWordScore uint8 // Blank=0, Yellow=1, Green=2; sum of first row
}

func AddWordleStats(m *discordgo.Message) (err error) {
	regex := regexp.MustCompile(`Wordle (\d*) (\w)\/6`)
	if !regex.MatchString(m.Content) {
		err := fmt.Errorf("message does not match wordle regex: \n%s", strings.ToLower(m.Content))
		log.Debug(err)
		return err
	}

	wordleStat := WordleStat{
		MessageID: m.ID,
		ChannelID: m.ChannelID,
		UserID:    m.Author.ID,
	}

	data := regex.FindStringSubmatch(m.Content)

	day, err := strconv.ParseInt(data[1], 10, 16)
	if err != nil {
		err := fmt.Errorf("error converting wordle day to int: %s", err.Error())
		log.Error(err)
		return err
	}
	wordleStat.Day = uint16(day)

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
	wordleStat.Score = uint8(score)

	// Skip the Wordle score line and the newline to find the actual squares and begin counting
	rows := strings.Split(m.Content, "\n")
	rows_of_squares := rows[2:]
	for i, row := range rows_of_squares {
		yellows := uint8(strings.Count(row, YELLOW_SQUARE_EMOJI))
		greens := uint8(strings.Count(row, GREEN_SQUARE_EMOJI))
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

	wordle, err := GetWordle(m.ChannelID)
	if err != nil {
		err := errors.New("no wordle game found in this channel")
		log.Debug(err)
		return err
	}

	err = s.MessageReactionAdd(m.ChannelID, m.ID, CALC_EMOJI)
	if err != nil {
		err := fmt.Errorf("unable to add reaction to wordle game results on messageid: %s\n%s", m.ID, err.Error())
		log.Error(err)
		return err
	}

	user := GetUser(m.Author)

	result = db.Limit(1).Where(&WordleStat{UserID: wordleStat.UserID}).Find(&existing)
	if result.RowsAffected == 0 {
		// User has never played
		wordle.Players = append(wordle.Players, &user)
		db.Save(&wordle)
	}

	db.Create(&wordleStat)
	user.CalculateStats()
	wordle.RefreshStatus()
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
			AddWordleStats(message)
		}
	}
}
