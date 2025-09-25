package utils

import "os"

// IsCIRun reports whether the current execution context is a CI run
// Will be deleted as CI env variable should be used
// For now preserved to be comliant with current proxy's and p2p ci
func IsCIRun() bool {
	return os.Getenv("CI_RUN") != ""
}
