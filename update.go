package main

import "github.com/hashicorp/go-version"

// Update - alter tables as needed
func (kdb *KDB) Update(ver string) {
	// 3.0.5 => 3.0.6
	// Add submitters to votes for further processing later
	curVer, err := version.NewVersion(ver)
	if err != nil {
		curVer, err = version.NewVersion("0")
		if err != nil {
			panic(err)
		}
	}

	v306, err := version.NewVersion("3.0.6")
	if err != nil {
		panic(err)
	}
	if curVer.LessThan(v306) {
		k.Log("KDB", "Attempting to update to DBv3.0.6")
		_, err = k.db.Exec(`ALTER TABLE votes ADD submitterID VARCHAR(32) NOT NULL`)
		if err != nil {
			k.Log("KDB", err.Error())
		}

		_, err = k.db.Exec(`DELETE FROM state`)
		if err != nil {
			panic(err)
		}

		_, err = k.db.Exec(`INSERT INTO state (version) VALUES(?)`, k.state.version)
		if err != nil {
			panic(err)
		}

	}

	// 3.0.6 => 3.0.7
	// Generalize id column to be used with users as well as messages
	v307, err := version.NewVersion("3.0.7")
	if err != nil {
		panic(err)
	}
	if curVer.LessThan(v307) {
		k.Log("KDB", "Attempting to update to DBv3.0.7")
		_, err = k.db.Exec(`ALTER TABLE watch CHANGE messageID id VARCHAR(64)`)
		if err != nil {
			k.Log("KDB", err.Error())
		}

		_, err = k.db.Exec(`DELETE FROM state`)
		if err != nil {
			panic(err)
		}

		_, err = k.db.Exec(`INSERT INTO state (version) VALUES(?)`, k.state.version)
		if err != nil {
			panic(err)
		}
	}

	k.Log("INFO", "Update to database successful, now version "+k.state.version)
}
