package logic

import (
	"strings"
)

func Trim(s string) string {
	return strings.Trim(s, "\"")
}
