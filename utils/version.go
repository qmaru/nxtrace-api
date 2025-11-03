package utils

import (
	"fmt"
	"sync"
)

const (
	DateVer   string = "COMMIT_DATE"
	CommitVer string = "COMMIT_VERSION"
	GoVer     string = "COMMIT_GOVER"
)

var Version = sync.OnceValue(func() string {
	return fmt.Sprintf("%s (git-%s) (%s)", DateVer, CommitVer, GoVer)
})
