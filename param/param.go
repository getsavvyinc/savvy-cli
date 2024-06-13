package param

import "regexp"

var paramRegex = regexp.MustCompile(`<([a-zA-Z0-9-_]+)>`)

func Extract(input string) []string {
	// Define a regular expression to match parameters in the form of <alphanumericchars>
	matches := paramRegex.FindAllStringSubmatch(input, -1)

	// Extract the matched parameters
	var params []string
	for _, match := range matches {
		if len(match) >= 1 {
			params = append(params, match[0])
		}
	}
	return params
}
