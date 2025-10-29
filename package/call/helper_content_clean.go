package call

import "strings"

func ContentClean(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimPrefix(content, "json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	return content
}
