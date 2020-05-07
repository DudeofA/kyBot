/* 	users.go
_________________________________
Holds all user functions for Kylixor Discord Bot
Andrew Langhill
kylixor.com
*/

package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

//----- U S E R   M A N A G E M E N T -----

// CreateUser - create user with default values and return it
func (kdb *KDB) CreateUser(s *discordgo.Session, id string) (user UserInfo) {

	// Get user info from discord
	discordUser, err := s.User(id)
	if err != nil {
		panic(err)
	}

	user.ID = id
	user.Name = discordUser.Username
	user.Discriminator = discordUser.Discriminator
	// Set defaults
	user.DoneDailies = false

	// Insert user into database
	_, err = k.db.Exec("INSERT INTO users (userID, name, discriminator, currentCID, lastSeenCID, credits, dailies) VALUES(?,?,?,?,?,?,?)",
		user.ID, user.Name, user.Discriminator, user.CurrentCID, user.LastSeenCID, user.Credits, user.DoneDailies)
	if err != nil {
		panic(err)
	}

	LogDB("User", user.Name, user.ID, "created in")

	return user
}

// ReadUser - Query database for user, creating a new one if none exists
func (kdb *KDB) ReadUser(s *discordgo.Session, userID string) (user UserInfo) {

	// Get user by userID
	row := k.db.QueryRow("SELECT userID, name, discriminator, currentCID, lastSeenCID, credits, dailies FROM users WHERE userID=(?)", userID)
	err := row.Scan(&user.ID, &user.Name, &user.Discriminator, &user.CurrentCID, &user.LastSeenCID, &user.Credits, &user.DoneDailies)
	switch err {
	case sql.ErrNoRows:
		k.Log("KDB", "User not found in DB, creating new...")
		return k.kdb.CreateUser(s, userID)
	case nil:
		LogDB("User", user.Name, user.ID, "read from")
		return user
	default:
		panic(err)
	}
}

// Update [User] - update user in database based on user argument
func (user *UserInfo) Update() {

	LogDB("User", user.Name, user.ID, "updated in")

	// Update the guild to the database
	_, err := k.db.Exec("INSERT INTO users (userID, name, discriminator, currentCID, lastSeenCID, credits, dailies) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE name = ?, discriminator = ?, currentCID = ?, lastSeenCID = ?, credits = ?, dailies = ?",
		user.ID, user.Name, user.Discriminator, user.CurrentCID, user.LastSeenCID, user.Credits, user.DoneDailies, user.Name, user.Discriminator, user.CurrentCID, user.LastSeenCID, user.Credits, user.DoneDailies)
	if err != nil {
		k.Log("FATAL", "Error updating user: "+user.Name)
		panic(err)
	}
}

// UpdateCredits - Changes the credits a user has, returning false if not enough
func (user *UserInfo) UpdateCredits(s *discordgo.Session, amt int) (success bool) {
	if user.Credits+amt > 0 {
		_, err := k.db.Exec("UPDATE users SET credits = ? WHERE userID = ?", user.Credits+amt, user.ID)
		if err != nil {
			panic(err)
		}
		return true
	}
	return false
}

// UpdateDailies [User] - update user dailies
func (user *UserInfo) UpdateDailies(s *discordgo.Session, arg bool) {
	user.DoneDailies = arg

	LogDB("User", user.Name, user.ID, fmt.Sprintf("dailiesDone->%t in", arg))

	// Update the guild to the database
	_, err := k.db.Exec("UPDATE users SET dailies = ? WHERE userID = ?", arg, user.ID)
	if err != nil {
		k.Log("FATAL", "Error updating user: "+user.Name)
		panic(err)
	}
}

// ResetDailies - Function to call once a day to reset dailies
func ResetDailies(s *discordgo.Session) {

	rows, err := k.db.Query("SELECT userID FROM users WHERE dailies > 0")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var user UserInfo
		if err := rows.Scan(&user.ID); err != nil {
			panic(err)
		}

		user = k.kdb.ReadUser(s, user.ID)
		user.UpdateDailies(s, false)
	}
}

// CompDailies - Gives everyone 2x daily amount who have coins to compensate for downtime
func CompDailies(s *discordgo.Session) {
	rows, err := k.db.Query("SELECT userID, credits FROM users WHERE credits > 0")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var user UserInfo
		if err := rows.Scan(&user.ID, &user.Credits); err != nil {
			panic(err)
		}

		user = k.kdb.ReadUser(s, user.ID)
		user.UpdateCredits(s, 2*k.botConfig.DailyAmt)
	}
}

// CollectDailies - Attempt to collect dailies
func (user *UserInfo) CollectDailies(s *discordgo.Session, m *discordgo.MessageCreate, msgGuild GuildInfo) {
	// If the dailies have not been done
	if !user.DoneDailies {
		// Mark dailies as done and add the appropriate amount
		user.UpdateDailies(s, true)
		user.Credits += k.botConfig.DailyAmt
		user.Update()
		// Indicate to user they have received their dailies
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
			"💵 | Daily %d coins received! Total %scoins: **%d**",
			k.botConfig.DailyAmt, msgGuild.Currency, user.Credits))
	} else {
		// Display time until dailies are available based on
		// when the next cronjob will run

		// _, nextRuntime := gocron.NextRun()
		jobs := k.cron.Entries()
		nextRuntime := jobs[0].Next
		timeUntil := time.Until(nextRuntime)
		hour := timeUntil / time.Hour
		timeUntil -= hour * time.Hour
		min := timeUntil / time.Minute
		timeUntil -= min * time.Minute
		sec := timeUntil / time.Second

		hourStr := "s"
		minStr := "s"
		secStr := "s"
		if hour == 1 {
			hourStr = ""
		}
		if min == 1 {
			minStr = ""
		}
		if sec == 1 {
			secStr = ""
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
			"💵 | You have already collected today's dailies.\nDailies reset in %d hour%s, %d minute%s and %d second%s.",
			hour, hourStr, min, minStr, sec, secStr))
	}
}
