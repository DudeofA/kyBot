package update

import (
	"kyBot/kyDB"
	"kyBot/status"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Temp to add previous users to Wordle object
func ConvertServerToWordle(s *discordgo.Session) {
	var server_objects []status.Server
	kyDB.DB.Where(&status.Server{Type: "wordle"}).Find(&server_objects)
	for _, server := range server_objects {
		wordle := &status.Wordle{
			ChannelID: server.StatusChannelID,
		}

		for _, id := range strings.Split(server.UserList, "\n") {
			var discord_user *discordgo.User
			var err error

			id = strings.Replace(id, "<", "", 1)
			id = strings.Replace(id, ">", "", 1)
			id = strings.Replace(id, "@", "", 1)
			id = strings.TrimSpace(id)

			if id != "" {
				user := &kyDB.User{
					ID: id,
				}
				discord_user, err = s.User(id)
				if err == nil {
					user.Username = discord_user.Username
					user.Discriminator = discord_user.Discriminator
				}
				var existing_user kyDB.User
				if result := kyDB.DB.Limit(1).Find(&existing_user, kyDB.User{ID: user.ID}); result.RowsAffected == 0 {
					kyDB.DB.Create(&user)
				}
				wordle.Users = append(wordle.Users, user)
			}
		}
		var existing_wordle *status.Wordle
		if result := kyDB.DB.Limit(1).Find(&existing_wordle, status.Wordle{ChannelID: wordle.ChannelID}); result.RowsAffected == 0 {
			kyDB.DB.Create(&wordle)
		}
		kyDB.DB.Delete(&server)
	}

	kyDB.DB.Migrator().DropColumn(&kyDB.User{}, "created_at")
	kyDB.DB.Migrator().DropColumn(&kyDB.User{}, "updated_at")
	kyDB.DB.Migrator().DropColumn(&kyDB.User{}, "deleted_at")
	kyDB.DB.Migrator().DropColumn(&kyDB.User{}, "name")
	kyDB.DB.Migrator().DropColumn(&kyDB.User{}, "current_vc_id")
	kyDB.DB.Migrator().DropColumn(&kyDB.User{}, "previous_vc_id")
	kyDB.DB.Migrator().DropColumn(&kyDB.User{}, "credits")
	kyDB.DB.Migrator().DropColumn(&kyDB.User{}, "got_dailies")

	kyDB.DB.Migrator().DropTable("guilds")
	kyDB.DB.Migrator().DropTable("hangmen")
	kyDB.DB.Migrator().DropTable("minecraft_servers")
}
