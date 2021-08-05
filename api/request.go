package api

import "strings"

func FacilitatorKey(headers map[string]string) string {
	for k, v := range headers {
		if strings.ToLower(k) == "x-facilitator-key" {
			return v
		}
	}
	return ""
}