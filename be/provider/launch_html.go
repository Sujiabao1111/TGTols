package provider

import (
	"encoding/json"
	"html"
	"net/url"
	"strings"
)

// NormalizeLaunchHTML decodes provider-returned HTML that was transported in a
// field normally used for URLs. It returns ok=false for normal launch URLs.
func NormalizeLaunchHTML(raw string) (string, bool) {
	candidates := []string{strings.TrimSpace(raw)}
	seen := map[string]struct{}{}

	for index := 0; index < len(candidates) && index < 24; index++ {
		current := strings.TrimSpace(candidates[index])
		if current == "" {
			continue
		}
		if _, exists := seen[current]; exists {
			continue
		}
		seen[current] = struct{}{}

		if looksLikeLaunchHTML(current) {
			return current, true
		}

		if decoded := html.UnescapeString(current); decoded != current {
			candidates = append(candidates, decoded)
		}
		if decoded, err := url.QueryUnescape(current); err == nil && decoded != current {
			candidates = append(candidates, decoded)
		}
		if decoded, ok := unescapeJSONString(current); ok && decoded != current {
			candidates = append(candidates, decoded)
		}
	}

	return "", false
}

func unescapeJSONString(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	var decoded string
	if strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`) {
		if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
			return decoded, true
		}
	}

	quoted := `"` + strings.ReplaceAll(trimmed, `"`, `\"`) + `"`
	if err := json.Unmarshal([]byte(quoted), &decoded); err != nil {
		return "", false
	}
	return decoded, true
}

func looksLikeLaunchHTML(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(normalized, "<!doctype") ||
		strings.Contains(normalized, "<html") ||
		strings.Contains(normalized, "<head") ||
		strings.Contains(normalized, "<body") ||
		strings.Contains(normalized, "<script") ||
		strings.Contains(normalized, "<iframe") ||
		strings.Contains(normalized, "<form")
}
