package rayiapp

import "github.com/sebarcode/codekit"

type ValidateJwtRequest struct {
	Token       string
	GetSessData bool
}

type ValidateJwtResponse struct {
	UserID      string
	SessionID   string
	ClientData  codekit.M
	SessionData codekit.M
}
