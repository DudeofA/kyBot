package component

import (
	"fmt"
	"kyBot/commands"
	"kyBot/kyDB"
	"sort"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

const (
	WORDLE_URL              = "https://www.powerlanguage.co.uk/wordle/"
	WORDLE_ROW_LENGTH       = 5
	WORDLE_GREEN_SQUARE     = "🟩"
	WORDLE_GREEN_SCORE      = 2
	WORDLE_YELLOW_SQUARE    = "🟨"
	WORDLE_YELLOW_SCORE     = 1
	WORDLE_COLOR            = 0x538d4e
	WORDLE_ACK_EMOJI        = "🧮"
	WORDLE_JOIN_EMOTE_NAME  = "aenezukojump"
	WORDLE_JOIN_EMOTE_ID    = "849514753042546719"
	WORDLE_LEAVE_EMOTE_NAME = "PES2_SadGeRain"
	WORDLE_LEAVE_EMOTE_ID   = "849698641869406261"
	WORDLE_FAIL_SCORE       = 7

	STOP_EMOJI = "🛑"
)

type Wordle struct {
	ChannelID       string       `gorm:"primaryKey"`
	Users           []*User      `gorm:"many2many:wordle_users;"`
	Stats           []WordleStat `gorm:"foreignKey:ChannelID"`
	StatusMessageID string
}

type WordlePlayerStats struct {
	User            *User
	AverageScore    float32
	AverageFirstRow float32
	GamesPlayed     int16
}

func init() {
	addServerCommand := &discordgo.ApplicationCommand{
		Name:        "add-wordle-channel",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Add a repeating wordle reminder message",
	}
	commands.AddCommand(addServerCommand)
}

func GetWordle(cid string) (wordle *Wordle, err error) {
	result := kyDB.DB.Preload(clause.Associations).Limit(1).Find(&wordle, Wordle{ChannelID: cid})
	if result.RowsAffected == 1 {
		return wordle, nil
	}
	return wordle, fmt.Errorf("no wordle found with channel id %s", cid)
}

func AddWordleChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "This channel will now track Wordle messages and stats",
		},
	}

	var wordle Wordle
	result := kyDB.DB.Take(&wordle, Wordle{ChannelID: i.Message.ChannelID})
	if result.RowsAffected != 0 {
		resp.Data.Content = fmt.Sprintf("Wordle channel already exists for this server: #%s", wordle.ChannelID)
		err := s.InteractionRespond(i.Interaction, resp)
		if err != nil {
			log.Errorf("Unable to respond to the interaction: %s", err.Error())
		}
		return
	}

	err := s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		log.Errorf("Unable to respond to the interaction: %s", err.Error())
	}

	wordle = Wordle{
		ChannelID: i.Message.ChannelID,
	}

	result = kyDB.DB.Create(&wordle)
	if result.Error != nil {
		log.Errorf("Unable to : %s", result.Error)
	}
}

func SendWordleReminders(s *discordgo.Session) {
	var wordles []Wordle
	kyDB.DB.Preload(clause.Associations).Find(&wordles)

	for _, wordle := range wordles {

		msg := wordle.buildEmbedMsg(s)
		update, err := s.ChannelMessageSendComplex(wordle.ChannelID, msg)
		if err != nil {
			log.Errorf("Error sending wordle update: %s", err.Error())
		}
		wordle.StatusMessageID = update.ID
		kyDB.DB.Where(&Wordle{ChannelID: wordle.ChannelID}).Updates(&Wordle{StatusMessageID: wordle.StatusMessageID})
	}
}

func (wordle *Wordle) AddUser(discord_user *discordgo.User) (changed bool) {
	for _, existingUser := range wordle.Users {
		if existingUser.ID == discord_user.ID {
			return false
		}
	}

	user := GetUser(discord_user)
	wordle.Users = append(wordle.Users, user)
	kyDB.DB.Model(&wordle).Association("Users").Append(&user)
	return true
}

func (wordle *Wordle) RemoveUser(discord_user *discordgo.User) (changed bool) {
	kyDB.DB.Find(&wordle, Wordle{ChannelID: wordle.ChannelID})
	changed = false
	i := 0
	for _, user := range wordle.Users {
		if user.ID != discord_user.ID {
			wordle.Users[i] = user
			i++
		} else {
			changed = true
		}
	}
	wordle.Users = wordle.Users[:i]
	kyDB.DB.Model(&wordle).Association("Users").Delete(&User{ID: discord_user.ID})
	return changed
}

func (wordle *Wordle) UpdateStatus(s *discordgo.Session) {
	msg := wordle.buildEmbedMsg(s)
	wordle.editStatusMessage(s, msg)
}

func (wordle *Wordle) buildEmbedMsg(s *discordgo.Session) (msg *discordgo.MessageSend) {
	optInButton := &discordgo.Button{
		Label: "Join Game",
		Style: 1,
		Emoji: discordgo.ComponentEmoji{
			Name:     WORDLE_JOIN_EMOTE_NAME,
			ID:       WORDLE_JOIN_EMOTE_ID,
			Animated: true,
		},
		CustomID: "join_wordle",
	}
	optOutButton := &discordgo.Button{
		Label: "Leave Game",
		Style: 4,
		Emoji: discordgo.ComponentEmoji{
			Name:     WORDLE_LEAVE_EMOTE_NAME,
			ID:       WORDLE_LEAVE_EMOTE_ID,
			Animated: true,
		},
		CustomID: "leave_wordle",
	}

	leaderBoard, worstFirstGuessUser := wordle.GenerateStatistics(s)
	var worstGuessUsername string
	var worstGuessValue float32
	if worstFirstGuessUser != nil {
		worstGuessUsername = worstFirstGuessUser.User.ID
		worstGuessValue = worstFirstGuessUser.AverageFirstRow
	} else {
		worstGuessUsername = "N/A"
		worstGuessValue = 0
	}

	wordleEmbed := &discordgo.MessageEmbed{
		URL:         WORDLE_URL,
		Title:       "CLICK HERE TO PLAY WORDLE",
		Description: "New wordle available now!",
		Timestamp:   "",
		Color:       WORDLE_COLOR,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Join In!",
				Value: "Click the button to join the game for tracking",
			},
			{
				Name:  "Worst First Guess",
				Value: fmt.Sprintf("<@%s> with average score of %.02f [Green=2,Yellow=1]", worstGuessUsername, worstGuessValue),
			},
			{
				Name:  "Leaderboard",
				Value: leaderBoard,
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

func (wordle *Wordle) editStatusMessage(s *discordgo.Session, updateContent *discordgo.MessageSend) {

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
	kyDB.DB.Where(&Wordle{ChannelID: wordle.ChannelID}).Updates(&Wordle{StatusMessageID: wordle.StatusMessageID})
}

func (wordle *Wordle) GenerateStatistics(s *discordgo.Session) (leaderBoard string, worstFirstRowUser *WordlePlayerStats) {
	leaderBoard = "None :("
	worstFirstRowUser = &WordlePlayerStats{
		User:            &User{Username: "No one :(", ID: "211307697331634186"},
		AverageScore:    0,
		AverageFirstRow: 0,
		GamesPlayed:     0,
	}

	if len(wordle.Users) != 0 {
		leaderBoard = "```\n Avg |Total| Name\n"
		var users []*WordlePlayerStats

		for _, player := range wordle.Users {
			if player.Username == "" {
				player.QueryInfo(s)
			}

			totalScore := int16(0)
			firstRowTotalScore := int16(0)
			gamesPlayed := int16(0)
			for _, stat := range wordle.Stats {
				if stat.UserID == player.ID {
					totalScore += int16(stat.Score)
					firstRowTotalScore += int16(stat.FirstWordScore)
					gamesPlayed++
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
			leaderBoard += fmt.Sprintf("%.2f | %03d | %s\n", user.AverageScore, user.GamesPlayed, user.User.Username)
		}
		leaderBoard += "\n```"
	}
	return leaderBoard, worstFirstRowUser
}
