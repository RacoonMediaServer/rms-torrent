package types

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
		return Error{Code: code, Underlying: err}
	}

	return nil
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
