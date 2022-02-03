package status

import (
	"fmt"
	"kyBot/commands"
	"kyBot/kyDB"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

	STOP_EMOJI = "🛑"
)

type Wordle struct {
	ChannelID       string       `gorm:"primaryKey"`
	Users           []*kyDB.User `gorm:"many2many:wordle_users;"`
	Stats           []WordleStat `gorm:"foreignKey:ChannelID"`
	StatusMessageID string
}

type WordleStat struct {
	gorm.Model
	MessageID      string
	UserID         string
	ChannelID      string
	Day            int16 // Wordle day number
	Score          int8  // Score out of 6
	BlankCount     int8  // Black or white squares
	YellowCount    int8
	GreenCount     int8
	FirstWordScore int8 // Blank=0, Yellow=1, Green=2; sum of first row
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

func AddWordleStats(s *discordgo.Session, m *discordgo.Message) (added bool) {
	regex := regexp.MustCompile(`Wordle (\d*) (\d)\/6`)
	if !regex.MatchString(m.Content) {
		log.Debug("Message does not match Wordle regex")
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
		if data[1] == "X" {
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

	wordle, err := GetWordle(m.ChannelID)
	if err != nil {
		log.Debug("No Wordle game found in this channel")
		return false
	}

	s.Ratelimiter.Lock()
	err = s.MessageReactionAdd(m.ChannelID, m.ID, WORDLE_ACK_EMOJI)
	if err != nil {
		log.Errorf("Unable to add reaction to Wordle game results: %s", err.Error())
		return false
	}
	s.Ratelimiter.Unlock()

	wordle.Stats = append(wordle.Stats, wordleStat)
	kyDB.DB.Model(&wordle).Where(&Wordle{ChannelID: wordle.ChannelID}).Updates(&wordle)
	return true
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
		kyDB.DB.Save(&wordle)
	}
}

func (wordle *Wordle) AddUser(discord_user *discordgo.User) (changed bool) {
	for _, existingUser := range wordle.Users {
		if existingUser.ID == discord_user.ID {
			return false
		}
	}

	user := kyDB.GetUser(discord_user)
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
	kyDB.DB.Model(&wordle).Association("Users").Delete(&kyDB.User{ID: discord_user.ID})
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

	var playerString string
	if len(wordle.Users) == 0 {
		playerString = "None :("
	} else {
		sortedUsers := wordle.Users
		sort.Slice(sortedUsers, func(i, j int) bool {
			return strings.ToLower(sortedUsers[i].Username) < strings.ToLower(sortedUsers[j].Username)
		})
		for _, player := range sortedUsers {
			if player.Username == "" {
				player.QueryInfo(s)
			}
			playerString += fmt.Sprintf("<@%s>\n", player.ID)

		}
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
				Name:  "Players",
				Value: playerString,
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
		// Send status
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
}

func (wordle *Wordle) CatchUp(s *discordgo.Session) {
	after := ""
	if len(wordle.Stats) > 0 {
		sort.Slice(wordle.Stats, func(i, j int) bool {
			return wordle.Stats[i].ID > wordle.Stats[j].ID
		})
		after = wordle.Stats[0].MessageID
	}
	messages, err := s.ChannelMessages(wordle.ChannelID, 100, "", after, "")
	if err != nil {
		log.Errorf("Unable to fetch channel messages from %s: %s", wordle.ChannelID, err.Error())
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})
	for _, message := range messages {
		if strings.HasPrefix(message.Content, "Wordle") {
			var wordle_stat *WordleStat
			if result := kyDB.DB.Limit(1).Find(&wordle_stat, WordleStat{MessageID: message.ID}); result.RowsAffected == 0 {
				AddWordleStats(s, message)
			} else {
				break
			}
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
			return wordle.Stats[i].ID < wordle.Stats[j].ID
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
