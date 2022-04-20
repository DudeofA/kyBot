package main

import log "github.com/sirupsen/logrus"

func CustomMigrateDB() {
	err := db.Migrator().AutoMigrate(&WordlePlayerStats{})
	if err != nil {
		log.Fatalf("Database migration failed: %s", err.Error())
	}

	var wordles []Wordle
	db.Preload("Remindees").Find(&wordles)
	for _, wordle := range wordles {
		var userIDs []string
		db.Model(&WordleStat{}).Where(&WordleStat{ChannelID: wordle.ChannelID}).Distinct("user_id").Find(&userIDs)
		for _, userID := range userIDs {
			var user User
			db.FirstOrCreate(&user, &User{ID: userID})
			db.Model(&wordle).Association("Players").Append(&user)

			var userStats WordlePlayerStats
			db.FirstOrCreate(&userStats, &WordlePlayerStats{UserID: userID})

			for _, remindee := range wordle.Remindees {
				if userID == remindee.ID {
					userStats.GetReminders = true
					db.Save(&userStats)
					break
				}
			}
			db.Save(&user)
		}
	}
}
