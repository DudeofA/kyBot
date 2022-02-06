package component

import (
	"kyBot/kyDB"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type User struct {
	ID            string `gorm:"primaryKey"` // User ID
	Username      string
	Discriminator string        // Unique identifier (#4712)
	Stats         []*WordleStat `gorm:"many2many:wordle_stats;"`
}

func GetUser(discord_user *discordgo.User) (user *User) {
	result := kyDB.DB.Limit(1).Find(&user, User{ID: discord_user.ID})
	if result.RowsAffected == 1 {
		return user
	}

	user = &User{
		ID:            discord_user.ID,
		Username:      discord_user.Username,
		Discriminator: discord_user.Discriminator,
	}
	kyDB.DB.Create(&user)

	return user
}

func (user *User) QueryInfo(s *discordgo.Session) {
	discord_user, err := s.User(user.ID)
	if err != nil {
		log.Errorf("Unable to get Discord user: %s", user.ID)
		return
	}
	user.Username = discord_user.Username
	user.Discriminator = discord_user.Discriminator
	kyDB.DB.Save(&user)
}

func (user *User) GetAverageScore() (average float32) {
	return 0
}
