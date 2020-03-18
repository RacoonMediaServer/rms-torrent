package types

import (
	"github.com/micro/go-micro/v2/logger"
)

const (
	AuthFailed = iota
	NetworkProblem
	CaptchaRequired
	UnknownTracker
	StorageProblem
)

type Error struct {
	Underlying error
	Code       int
	Captcha    string
}

func RaiseError(code int, err error) error {
	if err != nil {
		wrapped := Error{Code: code, Underlying: err}
		logger.Error(wrapped)
		return wrapped
	}

	return nil
}

func Raise(code int) error {
	err := Error{Code: code}
	logger.Error(err)
	return err
}

func (e Error) Error() string {
	message := ""
	switch e.Code {
	case AuthFailed:
		message = "Authentication failed"
	case NetworkProblem:
		message = "Network Problem"
	case CaptchaRequired:
		message = "Captcha required"
	default:
		if e.Underlying != nil {
			return e.Underlying.Error()
		}
	}

	if e.Underlying == nil {
		return message
	}

	return message + ": " + e.Underlying.Error()
}
