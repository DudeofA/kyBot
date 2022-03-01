package main

func MigrateDB() {
	db.Migrator().DropConstraint(&WordleStat{}, "fk_wordle_stats_wordle_stat")
	db.Migrator().DropColumn(&WordleStat{}, "created_at")
	db.Migrator().DropColumn(&WordleStat{}, "updated_at")
	db.Migrator().DropColumn(&WordleStat{}, "deleted_at")
	db.Migrator().DropColumn(&WordleStat{}, "wordle_stat_message_id")
}
