package info

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Printable() string {
	return fmt.Sprintf(
		`system-monitor [https://github.com/NicetasMatthias/SystemMonitor]
	Version = %s
	Commit = %s
	Build date = %s`, Version, Commit, Date)
}
