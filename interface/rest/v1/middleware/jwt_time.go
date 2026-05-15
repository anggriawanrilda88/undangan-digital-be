package middleware

import "time"

func jwtExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
}

func jwtNow() time.Time {
	return time.Now()
}
