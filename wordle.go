package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

const (
	WORDLE_URL          = "https://www.nytimes.com/games/wordle/index.html"
	WORDLE_ROW_LENGTH   = 5
	WORDLE_GREEN_SCORE  = 2
	WORDLE_YELLOW_SCORE = 1
	WORDLE_COLOR        = 0x538d4e
	WORDLE_FAIL_SCORE   = 7

	YELLOW_SQUARE_EMOJI = "🟨"
	GREEN_SQUARE_EMOJI  = "🟩"
	CALC_EMOJI          = "🧮"
	CHECK_EMOJI         = "✅"
	X_EMOJI             = "❌"
	STOP_EMOJI          = "🛑"
	BELL_EMOJI          = "🔔"
)

var WORDLE_DAY_0 = time.Date(2021, time.June, 19, 0, 0, 0, 0, time.Now().Location())

type Wordle struct {
	ChannelID       string `gorm:"primaryKey"`
	StatusMessageID string
	Remindees       []*User `gorm:"many2many:wordle_remindees"`
	Players         []*User `gorm:"many2many:wordle_players"`
}

type WordlePlayerStats struct {
	ID              uint `gorm:"primaryKey"`
	UserID          string
	AverageScore    float32
	AverageFirstRow float32
	GamesPlayed     uint16
	PlayedToday     bool
	GetReminders    bool
}

func GetWordle(channelID string) (wordle *Wordle, err error) {
	result := db.Preload("Players.WordleStats").Preload(clause.Associations).Limit(1).Find(&wordle, Wordle{ChannelID: channelID})
	if result.RowsAffected != 1 {
		return wordle, fmt.Errorf("no wordle found with channel id %s", channelID)
	}
	return wordle, nil
}

func WordleNewDay() {
	var wordles []Wordle
	db.Find(&wordles)

	for _, raw_wordle := range wordles {
		wordle, err := GetWordle(raw_wordle.ChannelID)
		if err != nil {
			log.Error(err)
		}
		wordle.StatusMessageID = ""
		wordle.RefreshStatus()
	}
}

func WordleSendReminder() {
	var wordles []Wordle
	db.Preload(clause.Associations).Find(&wordles)

	for _, wordle := range wordles {
		user_count := 0
		notification := "It's 7pm and you didn't do your Wordle yet :o\n"
		for _, user := range wordle.Remindees {
			var lastWordleStat WordleStat
			todayWordleDay := uint16(time.Since(WORDLE_DAY_0).Hours() / 24)
			db.Last(&lastWordleStat, &WordleStat{UserID: user.ID})
			if lastWordleStat.Day != todayWordleDay {
				notification += fmt.Sprintf("<@%s>\n", user.ID)
				user_count++
			}
		}
		if user_count > 0 {
			s.ChannelMessageSend(wordle.ChannelID, notification)
		}
	}
}

func (wordle *Wordle) RefreshStatus() {
	// TODO: Reload users more cleanly
	wordle, _ = GetWordle(wordle.ChannelID)
	msg := wordle.BuildEmbedMsg()
	wordle.EditStatusMessage(msg)
}

func (wordle *Wordle) BuildEmbedMsg() (msg *discordgo.MessageSend) {
	toggleReminderButton := &discordgo.Button{
		Label: "Toggle Reminders",
		Style: 1,
		Emoji: discordgo.ComponentEmoji{
			Name:     BELL_EMOJI,
			Animated: true,
		},
		CustomID: "toggle_wordle_reminder",
	}

	statusBoard := wordle.GenerateStatistics()

	wordleEmbed := &discordgo.MessageEmbed{
		URL:         WORDLE_URL,
		Title:       "CLICK HERE TO PLAY WORDLE",
		Description: "New wordle available now!",
		Timestamp:   "",
		Color:       WORDLE_COLOR,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Join In!",
				Value: "Click the button to be reminded if you haven't played by 7pm",
			},
			{
				Name:  "Statusboard",
				Value: statusBoard,
			},
		},
	}
	msg = &discordgo.MessageSend{
		Embed: wordleEmbed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					toggleReminderButton,
				},
			},
		},
	}
	return msg
}

func (wordle *Wordle) EditStatusMessage(updateContent *discordgo.MessageSend) {
	var statusMsg *discordgo.Message
	var err error

	_, err = s.ChannelMessage(wordle.ChannelID, wordle.StatusMessageID)
	if err != nil {
		// Send status if previous status message does not exist anymore (user deleted it)
		statusMsg, err = s.ChannelMessageSendComplex(wordle.ChannelID, updateContent)
		if err != nil {
			log.Errorln("Could not send server status message", err.Error())
			return
		}
	} else {
		// Edit existing message
		edit := &discordgo.MessageEdit{
			Components: updateContent.Components,
			ID:         wordle.StatusMessageID,
			Channel:    wordle.ChannelID,
			Embed:      updateContent.Embed,
		}
		statusMsg, err = s.ChannelMessageEditComplex(edit)
		if err != nil {
			log.Errorln("Could not edit wordle message", err.Error())
			return
		}
	}

	wordle.StatusMessageID = statusMsg.ID
	db.Where(&Wordle{ChannelID: wordle.ChannelID}).Updates(&Wordle{StatusMessageID: wordle.StatusMessageID})
}

func (wordle *Wordle) GenerateStatistics() (statusBoard string) {

	if len(wordle.Players) == 0 {
		return "No gamers :("
	}

	sort.Slice(wordle.Players, func(i, j int) bool {
		return wordle.Players[i].WordleStats.AverageScore < wordle.Players[j].WordleStats.AverageScore
	})

	statusBoard = "` Mean | Total | Today `\n"
	for _, user := range wordle.Players {
		playedStatus := X_EMOJI
		reminderStatus := ""

		if user.WordleStats.PlayedToday {
			playedStatus = CHECK_EMOJI
		}
		if user.WordleStats.GetReminders {
			reminderStatus = BELL_EMOJI
		}

		statusBoard += fmt.Sprintf(
			"` %.2f | %4d  |  %s  `%s<@%s>\n",
			user.WordleStats.AverageScore,
			user.WordleStats.GamesPlayed,
			playedStatus,
			reminderStatus,
			user.ID,
		)
	}
	return statusBoard
}
