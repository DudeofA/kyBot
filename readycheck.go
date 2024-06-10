package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	RED_SQUARE    = "🟥"
	YELLOW_SQUARE = "🟨"
	GREEN_SQUARE  = "🟩"
	BLACK_SQUARE  = "🔲"
)

type ReadyCheck struct {
	VoiceChannelID  string `gorm:"primaryKey"` // Voice Channel of ReadyCheck
	InitiatorUserID string // User ID of user that started ready check
	GuildID         string
	StartTime       time.Time
	UserCount       int    // Number of users being checked
	UserList        string // Comma delimited UserID list
	Complete        bool
	StatusChannelID string // Channel ID of the ready check
	StatusMessageID string // Message ID of the ready check response
	ReadyStatuses   []ReadyCheckStatus
}

type ReadyCheckStatus struct {
	ReadyCheckID string `gorm:"primaryKey"`
	UserID       string `gorm:"primaryKey"`
	Status       string // Status of user in ReadyCheck (Unknown, Ready, Not Ready)
}

func init() {
	addServerCommand := &discordgo.ApplicationCommand{
		Name:        "readycheck",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Do a ready check for all voice users",
	}
	AddCommand(addServerCommand)
}

func AddReadyCheck(i *discordgo.InteractionCreate) {
	var users []discordgo.User
	userVoiceChannel, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil {
		log.Errorln("Error fetching Voice State of interaction user: ", err)
		return
	}
	c, err := s.State.Channel(userVoiceChannel.ChannelID)
	if err != nil {
		log.Errorln("Error fetching channel of Voice State: ", err)
		return
	}
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		log.Errorln("Error fetching guild of Voice State channel: ", err)
		return
	}
	for _, vs := range g.VoiceStates {
		if vs.ChannelID == userVoiceChannel.ChannelID {
			user, err := s.State.Member(g.ID, vs.UserID)
			if err != nil {
				log.Errorln("Error fetching member of Voice State channel: ", err)
				return
			}
			users = append(users, *user.User)
		}
	}

	readyCheck := ReadyCheck{
		VoiceChannelID:  userVoiceChannel.ChannelID,
		InitiatorUserID: i.Member.User.ID,
		GuildID:         i.GuildID,
		StartTime:       time.Now(),
		UserCount:       len(users),
		Complete:        false,
		StatusChannelID: i.ChannelID,
	}

	if err := db.Take(&readyCheck).Error; err == nil {
		db.Delete(&readyCheck)
	}

	for _, user := range users {
		readyCheckStatus := ReadyCheckStatus{
			ReadyCheckID: readyCheck.VoiceChannelID,
			UserID:       user.ID,
			Status:       "Unknown",
		}
		readyCheck.ReadyStatuses = append(readyCheck.ReadyStatuses, readyCheckStatus)
	}
	db.Create(&readyCheck)
	msg := readyCheck.buildEmbedMsg()
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				msg.Embed,
			},
			Components: msg.Components,
		},
	}
	err = s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		log.Errorf("Error responding to the interaction: %s", err.Error())
		return
	}
	interaction, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.Errorf("Error getting interaction response message info: %s", err.Error())
		return
	}
	readyCheck.StatusMessageID = interaction.ID
	db.Save(&readyCheck)
}

func (readyCheck *ReadyCheck) Update() {
	db.Model(&readyCheck).Where(&ReadyCheck{VoiceChannelID: readyCheck.VoiceChannelID}).Updates(&readyCheck)
	msg := readyCheck.buildEmbedMsg()
	readyCheck.updateStatusMessage(msg)
}

func (readyCheck *ReadyCheck) Remove() {
	err := s.ChannelMessageDelete(readyCheck.StatusChannelID, readyCheck.StatusMessageID)
	if err != nil {
		log.Errorln("Error deleting Ready Check: ", err)
	}
	db.Select("ReadyStatuses").Delete(readyCheck)
}

func (readyCheck *ReadyCheck) updateStatusMessage(updateContent *discordgo.MessageSend) {

	var statusMsg *discordgo.Message
	var err error

	_, err = s.ChannelMessage(readyCheck.StatusChannelID, readyCheck.StatusMessageID)
	if err != nil {
		// Send status
		statusMsg, err = s.ChannelMessageSendComplex(readyCheck.StatusChannelID, updateContent)
		if err != nil {
			log.Errorln("Could not send server status message", err.Error())
			return
		}
	} else {
		// Edit existing message
		edit := &discordgo.MessageEdit{
			Components: &updateContent.Components,
			ID:         readyCheck.StatusMessageID,
			Channel:    readyCheck.StatusChannelID,
			Embed:      updateContent.Embed,
		}
		statusMsg, err = s.ChannelMessageEditComplex(edit)
		if err != nil {
			log.Errorln("Could not edit server message", err.Error())
			return
		}
	}

	readyCheck.StatusMessageID = statusMsg.ID
	db.Model(&readyCheck).Where(&ReadyCheck{VoiceChannelID: readyCheck.VoiceChannelID}).Updates(&Server{StatusMessageID: readyCheck.StatusMessageID})
}

func (readyCheck *ReadyCheck) buildEmbedMsg() (msg *discordgo.MessageSend) {

	var readyCheckList []string
	var userStatusList string
	for _, user := range readyCheck.ReadyStatuses {
		var square string
		switch user.Status {
		case "Unknown":
			square = YELLOW_SQUARE
		case "Ready":
			square = GREEN_SQUARE
		case "Not Ready":
			square = RED_SQUARE
		}
		readyCheckList = append(readyCheckList, square+square)
		discordUser, err := s.State.Member(readyCheck.GuildID, user.UserID)
		if err != nil {
			log.Errorln("Error retrieving user: ", err)
			return
		}
		userStatusList += fmt.Sprintf("%s: %s %s\n", discordUser.DisplayName(), square, user.Status)
	}
	readyCheckVisual := strings.Join(readyCheckList, BLACK_SQUARE)
	_ = readyCheckVisual

	color := 0xff0000
	status := "Active"
	if readyCheck.Complete {
		status = "Complete"
		color = 0x00ff00
	}
	statusEmbed := &discordgo.MessageEmbedField{
		Name:  status,
		Value: readyCheckVisual,
	}
	userEmbed := &discordgo.MessageEmbedField{
		Name:  "Users",
		Value: userStatusList,
	}

	title := "Ready Check"
	desc := fmt.Sprintf("Called by <@%s>", readyCheck.InitiatorUserID)
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       color,

		Fields: []*discordgo.MessageEmbedField{
			statusEmbed,
			userEmbed,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Started at: %s", time.Now().Local().Format("Jan 2 2006 - 3:04 PM MST")),
		},
	}

	ready_button := &discordgo.Button{
		Label:    "Ready",
		Style:    discordgo.SuccessButton,
		Disabled: false,
		CustomID: "ready",
	}
	not_ready_button := &discordgo.Button{
		Label:    "Not Ready",
		Style:    discordgo.DangerButton,
		Disabled: false,
		CustomID: "not_ready",
	}

	restart_button := &discordgo.Button{
		Label:    "Restart",
		Style:    discordgo.PrimaryButton,
		Disabled: true,
		CustomID: "restart_readycheck",
	}

	delete_button := &discordgo.Button{
		Label:    "Delete",
		Style:    discordgo.SecondaryButton,
		Disabled: false,
		CustomID: "delete_readycheck",
	}

	msg = &discordgo.MessageSend{
		Embed: embed,
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					ready_button,
					not_ready_button,
					restart_button,
					delete_button,
				},
			},
		},
	}
	return msg
}
