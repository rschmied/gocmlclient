package httputil

import "net/http"

// ApplyClientIdentityHeaders sets gocmlclient identity headers on h.
// Empty values are ignored.
func ApplyClientIdentityHeaders(h http.Header, clientID, clientUUID, clientVersion string) {
	if h == nil {
		return
	}
	if clientID != "" {
		h.Set("X-CML-CLIENT", clientID)
	}
	if clientVersion != "" {
		h.Set("X-CML-CLIENT-VERSION", clientVersion)
	}
	if clientUUID != "" {
		h.Set("X-Client-UUID", clientUUID)
	}
}
