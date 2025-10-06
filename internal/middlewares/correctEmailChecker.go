package middlewares

import "regexp"

func CorrectEmailChecker(email string) bool {
	const emailPattern = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	emailRegex := regexp.MustCompile(emailPattern)

	if emailRegex.MatchString(email) {
		return true
	}

	return false
}
