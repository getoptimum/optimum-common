package string

import (
	"regexp"
	stdstrings "strings"
)

var reg = regexp.MustCompile("[^a-zA-Z0-9]+")

// This useful utility is to be found at optimum-p2p/pkg/utils/strings.go
func SanitizeString(s string) string {
	s = stdstrings.TrimSpace(s)
	return reg.ReplaceAllString(stdstrings.ToLower(s), "")
}
