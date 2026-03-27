package http

import "time"

func parseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}
