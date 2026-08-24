package provider

import "strings"

func formatProviderReportTime(value string) string {
	return strings.TrimRight(value, "/")
}

func buildProviderReportURL(agentapi string, path string) string {
	return strings.TrimRight(agentapi, "/") + path
}

func isProviderReportSuccess(code int, message string) bool {
	trimmedMessage := strings.TrimSpace(strings.ToUpper(message))
	return code == 0 || code == 1 || trimmedMessage == "SUCCESS"
}
