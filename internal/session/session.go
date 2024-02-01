package session

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	sessions map[uuid.UUID]Session
)

func Init() {
	sessions = make(map[uuid.UUID]Session)
}

func CleanupTicker() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		<-ticker.C
		for _, s := range sessions {
			if time.Since(s.Created).Hours() >= 7*24 {
				DeleteSession(&s)
			}
		}
	}
}

type Session struct {
	ID          uuid.UUID
	Created     time.Time
	AccessToken string
	Login       string
	UserID      string
}

func NewSession(id uuid.UUID, accessToken string, login string, userID string) Session {
	return Session{
		ID:          id,
		Created:     time.Now(),
		AccessToken: accessToken,
		Login:       login,
		UserID:      userID,
	}
}

func AddSession(session Session) {
	sessions[session.ID] = session
}

func DeleteSession(session *Session) {
	delete(sessions, session.ID)
}

func SessionFromRequest(r *http.Request) (*Session, bool, error) {
	sessionCookie, err := r.Cookie("sessionid")
	if err != nil {
		return nil, false, nil
	}

	splits := strings.Split(sessionCookie.String(), "=")
	if len(splits) < 2 {
		return nil, false, errors.New("invalid cookie format")
	}

	sessionID, err := uuid.Parse(splits[1])
	if err != nil {
		return nil, false, err
	}

	s, ok := sessions[sessionID]
	if !ok {
		return nil, false, nil
	}

	return &s, true, nil
}
