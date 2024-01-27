package session

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var (
	// TODO: expire sessions after a week (like cookies)
	sessions map[uuid.UUID]Session
)

func Init() {
	sessions = make(map[uuid.UUID]Session)
}

type Session struct {
	Id          uuid.UUID
	AccessToken string
	Login       string
	UserId      string
}

func NewSession(id uuid.UUID, accessToken string, login string, userId string) Session {
	return Session{
		Id:          id,
		AccessToken: accessToken,
		Login:       login,
		UserId:      userId,
	}
}

func AddSession(session Session) {
	sessions[session.Id] = session
}

func DeleteSession(session *Session) {
	delete(sessions, session.Id)
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

	sessionId, err := uuid.Parse(splits[1])
	if err != nil {
		return nil, false, err
	}

	s, ok := sessions[sessionId]
	if !ok {
		return nil, false, nil
	}

	return &s, true, nil
}
