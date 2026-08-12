package rayiapp

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

const (
	AuthProviderResultRedirect = "redirect"
	AuthProviderResultIdentity = "identity"
)

type AuthProviderAuthenticateRequest struct {
	ProviderID   string
	ProviderType string
	ProviderName string
	ProviderData codekit.M
	Payload      codekit.M
	Query        codekit.M
	Headers      codekit.M
}

type AuthProviderAuthenticateResult struct {
	ResultType  string
	RedirectURL string
	StatusCode  int
	UserID      string
	LoginID     string
	Email       string
	Name        string
	Realm       string
	ReferenceID string
	SessionData codekit.M
	ClientData  codekit.M
	Config      codekit.M
}

type AuthProviderAuthenticator interface {
	Authenticate(ctx *kaos.Context, req *AuthProviderAuthenticateRequest) (*AuthProviderAuthenticateResult, error)
}

var (
	authProviderMu             sync.RWMutex
	authProviderAuthenticators = map[string]AuthProviderAuthenticator{}
)

func RegisterAuthProviderAuthenticator(providerType string, authenticator AuthProviderAuthenticator) {
	providerType = strings.TrimSpace(strings.ToLower(providerType))
	if providerType == "" || authenticator == nil {
		return
	}

	authProviderMu.Lock()
	authProviderAuthenticators[providerType] = authenticator
	authProviderMu.Unlock()
}

func GetAuthProviderAuthenticator(providerType string) (AuthProviderAuthenticator, bool) {
	authProviderMu.RLock()
	authenticator, ok := authProviderAuthenticators[strings.TrimSpace(strings.ToLower(providerType))]
	authProviderMu.RUnlock()
	return authenticator, ok
}

func StopHTTPResponse(ctx *kaos.Context) {
	if ctx == nil {
		return
	}
	ctx.Data().Set("kaos_command_1", "stop")
}

func Redirect(ctx *kaos.Context, targetURL string, statusCode int) error {
	if ctx == nil {
		return fmt.Errorf("missing context")
	}
	if targetURL == "" {
		return fmt.Errorf("missing redirect target")
	}

	writer := ctx.HttpWriter()
	if writer == nil {
		return fmt.Errorf("missing http writer")
	}

	request := ctx.HttpRequest()
	if request == nil {
		return fmt.Errorf("missing http request")
	}

	if statusCode == 0 {
		statusCode = http.StatusFound
	}

	http.Redirect(writer, request, targetURL, statusCode)
	StopHTTPResponse(ctx)
	return nil
}
