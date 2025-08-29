package rayiapp

import (
	"net/http"

	"git.kanosolution.net/kano/kaos"
)

const (
	CtxJWTReferenceID = "jwt_reference_id"
	HTTP_REQUEST      = "http_request"
	HTTP_WRITER       = "http_writer"
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
		opts = &kaos.PublishOpts{}
	}

	ctxData := ctx.Data().Data()
	for _, dataName := range dataNames {
		v, ok := ctxData[dataName]
		if ok {
			opts.Headers.Set(dataName, v)
		}
	}

	return opts
}
