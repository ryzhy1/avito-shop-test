package middlewares

func IdentifyLoginInputType(input string) string {
	if CorrectEmailChecker(input) {
		return "email"
	}
	return "username"
}
