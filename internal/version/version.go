package version

import "strings"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type Info struct {
	Version string
	Commit  string
	Date    string
}

func BuildInfo() Info {
	return Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}

func (i Info) String() string {
	parts := []string{i.Version}
	if i.Commit != "" && i.Commit != "unknown" {
		parts = append(parts, i.Commit)
	}
	if i.Date != "" && i.Date != "unknown" {
		parts = append(parts, i.Date)
	}
	return strings.Join(parts, " ")
}
