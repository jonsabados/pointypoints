package cors

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
)

type ResponseHeaderBuilder func(ctx context.Context, inboundHeaders map[string]string) map[string]string

func NewResponseHeaderBuilder(allowedDomains []string) ResponseHeaderBuilder {
	return func(ctx context.Context, inboundHeaders map[string]string) map[string]string {
		origin := ""
		for k, v := range inboundHeaders {
			if strings.ToLower(k) == "origin" {
				origin = v
				break
			}
		}
		headers := make(map[string]string)
		if origin != "" && isOriginAllowed(origin, allowedDomains) {
			headers["Access-Control-Allow-Origin"] = origin
			headers["Access-Control-Allow-Headers"] = "Authorization,Content-Type,X-Facilitator-Key"
			headers["Access-Control-Expose-Headers"] = "Location"
			headers["Access-Control-Allow-Methods"] = "OPTIONS,HEAD,GET,POST,PUT,DELETE"
		} else {
			zerolog.Ctx(ctx).Warn().Interface("allowedDomains", allowedDomains).Str("origin", origin).Msg("disallowed origin")
		}
		headers["Vary"] = "Origin"
		return headers
	}
}

func isOriginAllowed(origin string, allowedDomains []string) bool {
	for _, o := range allowedDomains {
		if origin == o {
			return true
		}
	}
	return false
}