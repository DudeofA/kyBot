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
	WORDLE_URL              = "https://www.nytimes.com/games/wordle/index.html"
	WORDLE_ROW_LENGTH       = 5
	WORDLE_GREEN_SCORE      = 2
	WORDLE_YELLOW_SCORE     = 1
	WORDLE_COLOR            = 0x538d4e
	WORDLE_JOIN_EMOTE_NAME  = "aenezukojump"
	WORDLE_JOIN_EMOTE_ID    = "849514753042546719"
	WORDLE_LEAVE_EMOTE_NAME = "PES2_SadGeRain"
	WORDLE_LEAVE_EMOTE_ID   = "849698641869406261"
	WORDLE_FAIL_SCORE       = 7

	YELLOW_SQUARE_EMOJI = "🟨"
	GREEN_SQUARE_EMOJI  = "🟩"
	CALC_EMOJI          = "🧮"
	CHECK_EMOJI         = "✅"
	X_EMOJI             = "❌"
	STOP_EMOJI          = "🛑"
	TIME_EMOJI          = "⌛"
)

var WORDLE_DAY_0 = time.Date(2021, time.June, 19, 0, 0, 0, 0, time.Now().Location())

type Wordle struct {
	ChannelID       string `gorm:"primaryKey"`
	StatusMessageID string
	Remindees       []*User `gorm:"many2many:wordle_remindees"`
	Players         []*User `gorm:"many2many:wordle_players"`
}

type WordlePlayerStats struct {
	User            *User
	AverageScore    float32
	AverageFirstRow float32
	GamesPlayed     int16
	PlayedToday     bool
}

func GetWordle(channelID string) (wordle Wordle, err error) {
	result := db.Preload(clause.Associations).Limit(1).Find(&wordle, Wordle{ChannelID: channelID})
	if result.RowsAffected != 1 {
		return wordle, fmt.Errorf("no wordle found with channel id %s", channelID)
	}
	return wordle, nil
}

func EnableWordleReminder(i *discordgo.InteractionCreate) (err error) {
	wordle, err := GetWordle(i.Message.ChannelID)
	if err != nil {
		return err
	}

	for _, existingUser := range wordle.Remindees {
		if existingUser.ID == i.Member.User.ID {
			return nil // No action needed if users is already a part of the notifications group
		}
	}

	user := GetUser(i.Member.User)
	wordle.Remindees = append(wordle.Remindees, &user)

	wordle.UpdateStatus()
	db.Save(&wordle)
	return nil
}

func DisableWordleReminder(i *discordgo.InteractionCreate) (err error) {
	wordle, err := GetWordle(i.Message.ChannelID)
	if err != nil {
		return err
	}

	db.Model(&wordle).Association("Remindees").Delete(&User{ID: i.Member.User.ID})

	wordle.UpdateStatus()
	db.Save(&wordle)
	return nil
}

func WordleNewDay() {
	var wordles []Wordle
	db.Preload(clause.Associations).Find(&wordles)

	for _, wordle := range wordles {
		wordle.StatusMessageID = ""
		wordle.UpdateStatus()
	}
}

func WordleSendReminder() {
	var wordles []Wordle
	db.Preload(clause.Associations).Find(&wordles)

	for _, wordle := range wordles {
		notification := "It's 7pm and you didn't do your Wordle yet :o\n"
		for _, user := range wordle.Remindees {
			var lastWordleStat WordleStat
			todayWordleDay := int16(time.Since(WORDLE_DAY_0).Hours() / 24)
			db.Last(&lastWordleStat, &WordleStat{UserID: user.ID})
			if lastWordleStat.Day != todayWordleDay {
				notification += fmt.Sprintf("<@%s>\n", user.ID)
			}
		}
		s.ChannelMessageSend(wordle.ChannelID, notification)
	}
}

func (wordle *Wordle) UpdateStatus() {
	msg := wordle.BuildEmbedMsg()
	wordle.EditStatusMessage(msg)
}

func (wordle *Wordle) BuildEmbedMsg() (msg *discordgo.MessageSend) {
	optInButton := &discordgo.Button{
		Label: "Get Reminders",
		Style: 1,
		Emoji: discordgo.ComponentEmoji{
			Name:     TIME_EMOJI,
			Animated: true,
		},
		CustomID: "enable_wordle_reminder",
	}
	optOutButton := &discordgo.Button{
		Label: "Stop Reminders",
		Style: 4,
		Emoji: discordgo.ComponentEmoji{
			Name:     STOP_EMOJI,
			Animated: true,
		},
		CustomID: "disable_wordle_reminder",
	}

	leaderboard, worstFirstGuessUser := wordle.GenerateStatistics()
	var worstGuessUsername string
	var worstGuessValue float32
	if worstFirstGuessUser != nil {
		worstGuessUsername = worstFirstGuessUser.User.ID
		worstGuessValue = worstFirstGuessUser.AverageFirstRow
	} else {
		worstGuessUsername = "N/A"
		worstGuessValue = 0
	}

	remindersString := "```\n"
	for _, user := range wordle.Remindees {
		remindersString += user.Username + "\n"
	}
	remindersString += "\n```"

	reminders := remindersString

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
				Name:  "Leaderboard",
				Value: leaderboard,
			},
			{
				Name:  "Worst First Guess",
				Value: fmt.Sprintf("<@%s> with average score of %.02f [Green=2,Yellow=1]", worstGuessUsername, worstGuessValue),
			},
			{
				Name:  "People to Remind",
				Value: reminders,
			},
		},
	}
	msg = &discordgo.MessageSend{
		Embed: wordleEmbed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					optInButton,
					optOutButton,
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
		// Send status if previous status message does not exist anymore (user deleted)
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

func (wordle *Wordle) GenerateStatistics() (leaderBoard string, worstFirstRowUser *WordlePlayerStats) {
	leaderBoard = "None :("
	worstFirstRowUser = &WordlePlayerStats{
		User:            &User{Username: "No one :(", ID: "211307697331634186"},
		AverageScore:    0,
		AverageFirstRow: 0,
		GamesPlayed:     0,
	}

	if len(wordle.Players) != 0 {
		leaderBoard = "` Mean | Total | Today `\n"
		var users []*WordlePlayerStats

		// if # players != # unique user ids of wordle stats, rebuild user list
		var userIDs []string
		db.Model(&WordleStat{}).Distinct("user_id").Find(&userIDs)
		if len(wordle.Players) != len(userIDs) {
			wordle.Players = nil
			for _, userID := range userIDs {
				var user User
				db.Take(&user, &User{ID: userID})
				wordle.Players = append(wordle.Players, &user)
			}
			db.Save(&wordle)
		}

		for _, player := range wordle.Players {
			if player.Username == "" {
				player.QueryInfo()
			}

			totalScore := int16(0)
			firstRowTotalScore := int16(0)
			gamesPlayed := int16(0)
			playedToday := false
			var stats []WordleStat
			db.Find(&stats, WordleStat{ChannelID: wordle.ChannelID, UserID: player.ID})
			for _, stat := range stats {
				totalScore += int16(stat.Score)
				firstRowTotalScore += int16(stat.FirstWordScore)
				gamesPlayed++
				todayWordleDay := int16(time.Since(WORDLE_DAY_0).Hours() / 24)
				if stat.Day == todayWordleDay {
					playedToday = true
				}
			}

			var averageScore float32
			var averageFirstRowScore float32
			if gamesPlayed == 0 {
				averageScore = 7
				averageFirstRowScore = 7
			} else {
				averageScore = float32(totalScore) / float32(gamesPlayed)
				averageFirstRowScore = float32(firstRowTotalScore) / float32(gamesPlayed)
			}

			user := &WordlePlayerStats{
				User:            player,
				AverageScore:    averageScore,
				AverageFirstRow: averageFirstRowScore,
				GamesPlayed:     gamesPlayed,
				PlayedToday:     playedToday,
			}
			users = append(users, user)

			if worstFirstRowUser.AverageFirstRow == 0 || user.AverageFirstRow < worstFirstRowUser.AverageFirstRow {
				worstFirstRowUser = user
			}
		}
		sort.Slice(users, func(i, j int) bool {
			return users[i].AverageScore < users[j].AverageScore
		})
		for _, user := range users {
			playedStatus := X_EMOJI
			if user.PlayedToday {
				playedStatus = CHECK_EMOJI
			}
			leaderBoard += fmt.Sprintf("` %.2f | %4d  |  %s  `  <@%s>\n", user.AverageScore, user.GamesPlayed, playedStatus, user.User.ID)
		}
	}
	return leaderBoard, worstFirstRowUser
}
