package auth

import (
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) string {
	key := headers.Get("Authorization")
	key = strings.TrimPrefix(key, "ApiKey ")

	return key
}
