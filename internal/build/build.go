package build

import "time"

var Version = "DEV"

var Date = "" // YYYY-MM-DD

func init() {
	if Version == "DEV" {
		Date = time.Now().Format("2006-01-02")
	}
}
