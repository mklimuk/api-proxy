package proxy

import "strings"

func extractToken(tokenString string) string {
	return strings.TrimPrefix(tokenString, "Bearer ")
}
