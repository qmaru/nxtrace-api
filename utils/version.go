package utils

import (
	"fmt"
)

const (
	Major     string = "1"
	Minor     string = "0"
	DateVer   string = "COMMIT_DATE"
	CommitVer string = "COMMIT_VERSION"
	GoVer     string = "COMMIT_GOVER"
)

var Version string = fmt.Sprintf("%s.%s-%s (git-%s) (%s)", Major, Minor, DateVer, CommitVer, GoVer)
