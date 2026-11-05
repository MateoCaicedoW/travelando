package helpers

import (
	"errors"
	"net/http"

	"go.leapkit.dev/core/server/session"
)

const (
	FlashError   = "error"
	FlashSuccess = "success"

	flashMessageKey = "travelando_flash_message"
	flashLevelKey   = "travelando_flash_level"
)

var errSessionUnavailable = errors.New("session unavailable")

type Flash struct {
	Message string
	Level   string
}

func SetFlash(r *http.Request, level, message string) error {
	sess := session.FromCtx(r.Context())
	if sess == nil {
		return errSessionUnavailable
	}
	sess.Values[flashMessageKey] = message
	sess.Values[flashLevelKey] = level
	return nil
}

func ConsumeFlash(r *http.Request) Flash {
	if r == nil {
		return Flash{}
	}
	sess := session.FromCtx(r.Context())
	if sess == nil {
		return Flash{}
	}
	message, messageOK := sess.Values[flashMessageKey].(string)
	level, levelOK := sess.Values[flashLevelKey].(string)
	delete(sess.Values, flashMessageKey)
	delete(sess.Values, flashLevelKey)
	if !messageOK || !levelOK {
		return Flash{}
	}
	return Flash{Message: message, Level: level}
}
