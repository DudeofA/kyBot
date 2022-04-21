package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type User struct {
	ID            string `gorm:"primaryKey"` // Discord User ID
	Username      string
	Discriminator string // Unique identifier (#4712)
	WordleGames   []*WordleStat
	WordleStats   *WordlePlayerStats
}

func GetUser(discord_user *discordgo.User) (user User) {
	result := db.Preload("WordleStats").Limit(1).Find(&user, User{ID: discord_user.ID})
	if result.RowsAffected == 1 {
		return user
	}

	user = User{
		ID:            discord_user.ID,
		Username:      discord_user.Username,
		Discriminator: discord_user.Discriminator,
	}
	db.Create(&user)
	return user
}

func (user *User) QueryInfo() {
	var discord_user *discordgo.User
	var err error
	if user.Username == "" {
		discord_user, err = s.User(user.ID)
		if err != nil {
			log.Errorf("Unable to get Discord user: %s", user.ID)
			return
		}
		user.Username = discord_user.Username
		user.Discriminator = discord_user.Discriminator
	}

	db.Save(&user)
}

func (user *User) CalculateStats() {
	if user.WordleStats == nil {
		user.WordleStats = &WordlePlayerStats{
			GetReminders: false,
		}
	}

	user.WordleStats.AverageScore = user.GetAverageScore()
	user.WordleStats.AverageFirstRow = user.GetAverageFirstRow()
	user.WordleStats.GamesPlayed = user.GetGamesPlayed()
	user.WordleStats.PlayedToday = user.CheckPlayedToday()

	db.Save(&user.WordleStats)
}

func (user *User) GetAverageScore() (average float32) {
	row := db.Model(&WordleStat{}).Where(&WordleStat{UserID: user.ID}).Select("avg(score)").Row()
	row.Scan(&average)
	return average
}

func (user *User) GetAverageFirstRow() (average float32) {
	row := db.Model(&WordleStat{}).Where(&WordleStat{UserID: user.ID}).Select("avg(first_word_score)").Row()
	row.Scan(&average)
	return average
}

func (user *User) GetGamesPlayed() (count uint16) {
	var temp_count int64
	db.Model(&WordleStat{}).Where(&WordleStat{UserID: user.ID}).Count(&temp_count)
	count = uint16(temp_count)
	return count
}

func (user *User) CheckPlayedToday() bool {
	todayWordleDay := uint16(time.Since(WORDLE_DAY_0).Hours() / 24)
	var stats []WordleStat
	db.Find(&stats, WordleStat{UserID: user.ID})
	for _, stat := range stats {
		if stat.Day == todayWordleDay {
			return true
		}
	}
	return false
}

func (user *User) ToggleWordleReminder() error {
	var stat WordlePlayerStats
	result := db.FirstOrCreate(&stat, &WordlePlayerStats{UserID: user.ID})
	if result.Error != nil {
		return result.Error
	}
	if stat.GetReminders {
		stat.GetReminders = false
	} else {
		stat.GetReminders = true
	}
	db.Save(&stat)

	return nil
}
