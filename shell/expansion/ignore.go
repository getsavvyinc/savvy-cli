package expansion

import "regexp"

var grepColorExcludeDirRegexPattern = regexp.MustCompile(`grep --color=auto --exclude-dir={[\w,.]+}`)

func IgnoreGrep(message string) string {
	// grep expansions are not useful for the user with the addition of --color=auto --exclude-dir={.bzr,CVS,.git,.hg,.svn} by default.
	// This is a workaround to ignore the expansions in the command.
	if grepColorExcludeDirRegexPattern.MatchString(message) {
		return grepColorExcludeDirRegexPattern.ReplaceAllString(message, "grep")
	}
	return message
}
