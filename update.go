package main

import "github.com/hashicorp/go-version"

// Update - alter tables as needed
func (kdb *KDB) Update(ver string) {
	// 3.0.5 => 3.0.6
	curVer, err := version.NewVersion(ver)
	if err != nil {
		curVer, err = version.NewVersion("0")
		if err != nil {
			panic(err)
		}
	}
	newVer, err := version.NewVersion("3.0.6")
	if err != nil {
		panic(err)
	}

	if curVer.LessThan(newVer) {
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

		k.Log("INFO", "Update to database successful, now version "+k.state.version)
	}

}
