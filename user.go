package main

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type User struct {
	ID            string `gorm:"primaryKey"` // User ID
	Username      string
	Discriminator string // Unique identifier (#4712)
	Stats         []*WordleStat
}

func GetUser(discord_user *discordgo.User) (user User) {
	result := db.Limit(1).Find(&user, User{ID: discord_user.ID})
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

func (user *User) GetAverageScore() (average float32) {
	row := db.Model(&WordleStat{}).Where(&WordleStat{UserID: user.ID}).Select("avg(score)").Row()
	row.Scan(&average)
	return average
}

func (user *User) GetGamesPlayed() (count int16) {
	var temp_count int64
	db.Model(&WordleStat{}).Where(&WordleStat{UserID: user.ID}).Count(&temp_count)
	count = int16(temp_count)
	return count
}
