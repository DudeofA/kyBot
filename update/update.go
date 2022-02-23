package update

import (
	"kyBot/component"
	"kyBot/kyDB"
)

func Update() {
	kyDB.DB.Migrator().DropColumn(&component.WordleStat{}, "created_at")
	kyDB.DB.Migrator().DropColumn(&component.WordleStat{}, "updated_at")
	kyDB.DB.Migrator().DropColumn(&component.WordleStat{}, "deleted_at")
	kyDB.DB.Migrator().DropColumn(&component.WordleStat{}, "wordle_stat_message_id")
}
