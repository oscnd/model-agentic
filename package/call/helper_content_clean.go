package call

import "strings"

func ContentClean(content string) string {
	bracketIndex := strings.Index(content, "{")
	squareBracketIndex := strings.Index(content, "[")
	if bracketIndex != -1 && (squareBracketIndex == -1 || bracketIndex < squareBracketIndex) {
		content = content[bracketIndex:]
		endIndex := strings.LastIndex(content, "}")
		if endIndex != -1 {
			content = content[:endIndex+1]
		}
	}
	if squareBracketIndex != -1 && (bracketIndex == -1 || squareBracketIndex < bracketIndex) {
		content = content[squareBracketIndex:]
		endIndex := strings.LastIndex(content, "]")
		if endIndex != -1 {
			content = content[:endIndex+1]
		}
	}

	return content
}
