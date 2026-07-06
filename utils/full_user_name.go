package utils

import (
	"strings"

	"github.com/gotd/td/tg"
)

func FullUserName(user *tg.User) string {
	if user == nil {
		return ""
	}

	parts := []string{}
	firstName, ok := user.GetFirstName()
	if ok {
		parts = append(parts, firstName)
	}
	lastName, ok := user.GetLastName()
	if ok {
		parts = append(parts, lastName)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
