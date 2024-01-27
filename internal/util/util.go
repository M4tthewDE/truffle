package util

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func SessionIdFromRequest(r *http.Request) (*uuid.UUID, error) {
	sessionCookie, err := r.Cookie("sessionid")
	if err != nil {
		return nil, err
	}

	splits := strings.Split(sessionCookie.String(), "=")
	if len(splits) < 2 {
		return nil, errors.New("invalid cookie format")
	}

	sessionId, err := uuid.Parse(splits[1])
	if err != nil {
		return nil, err
	}

	return &sessionId, nil
}
