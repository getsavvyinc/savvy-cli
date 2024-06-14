package param

import "regexp"

var paramRegex = regexp.MustCompile(`<([a-zA-Z0-9-_]+)>`)

func Extract(input string) []string {
	// Define a regular expression to match parameters in the form of <alphanumericchars>
	matches := paramRegex.FindAllStringSubmatch(input, -1)

	seen := make(map[string]struct{})
	// Extract the matched parameters
	var params []string
	for _, match := range matches {
		if len(match) >= 1 {
			if _, ok := seen[match[0]]; ok {
				continue
			}
			seen[match[0]] = struct{}{}
			params = append(params, match[0])
		}
	}
	return params
}
