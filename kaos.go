package rayiapp

import (
	"net/http"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

const (
	CtxJwtToken       = "jwt_token"
	CtxJWTReferenceID = "jwt_reference_id"
	CtxJwtSessionID   = "jwt_sess_id"
	CtxJwtReferenceID = "jwt_reference_id"
	CtxJwtSessionData = "jwt_sess_data"
	CtxJwtClientData  = "jwt_client_data"

	HTTP_REQUEST = "http_request"
	HTTP_WRITER  = "http_writer"
)

// GetAccountID from given kaos context
func GetUserIDFromCtx(ctx *kaos.Context) string {
	return ctx.Data().Get(CtxJWTReferenceID, "").(string)
}

// GetHTTPRequest from given kaos context
func GetHTTPRequest(ctx *kaos.Context) (*http.Request, bool) {
	hr, ok := ctx.Data().Get(HTTP_REQUEST, nil).(*http.Request)
	return hr, ok
}

// CopyContextDataToPublishOptions copy context data to publish options. It is useful to pass data to request between microservices
func CopyContextDataToPublishOptions(ctx *kaos.Context, opts *kaos.PublishOpts, dataNames ...string) *kaos.PublishOpts {
	if len(dataNames) == 0 {
		dataNames = ctx.Data().Keys()
	}

	if opts == nil {
		opts = new(kaos.PublishOpts)
	}
	if opts.Headers == nil {
		opts.Headers = codekit.M{}
	}
	if opts.Config == nil {
		opts.Config = codekit.M{}
	}

	ctxData := ctx.Data().Data()
	for _, dataName := range dataNames {
		v, ok := ctxData[dataName]
		if ok {
			opts.Headers.Set(dataName, v)
		}
	}
	if tenantID := GetTenantID(ctx); tenantID != "" {
		// HTEV only forwards string-valued headers. Flatten TenantID from the
		// jwt_client_data map so tenant context survives service-to-service calls.
		opts.Headers.Set("TenantID", tenantID)
	}

	return opts
}
