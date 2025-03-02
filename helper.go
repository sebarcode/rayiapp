package rayiapp

import (
	"fmt"
	"net/http"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
	"github.com/spf13/viper"
)

const (
	HttpWriter     string = "http_writer"
	HttpRequest    string = "http_request"
	JwtReferenceID string = "jwt_reference_id"
	JwtData        string = "jwt_data"
)

func ReadConfig(configPath string, dest interface{}) error {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("unable to read config file. %s", err.Error())
	}

	if err := v.Unmarshal(dest); err != nil {
		return fmt.Errorf("unable to parse config file. %s", err.Error())
	}

	return nil
}

func WrapApiError(ctx *kaos.Context, errTxt string) {
	w := ctx.HttpWriter()
	if w == nil {
		ctx.Log().Errorf("http writer is nil")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(codekit.Jsonify(codekit.M{}.Set("error", errTxt).Set("timestamp", time.Now())))

	r := ctx.HttpRequest()
	if r != nil {
		ctx.Log().Errorf("%s | %s", r.URL.String(), errTxt)
	}
}

func PrepareCtxData(ctx *kaos.Context, userid, companyid string) *kaos.Context {
	if userid != "" {
		ctx.Data().Set(JwtReferenceID, userid)
	}

	if companyid == "" {
		companyid = "Demo00"
	}
	ctx.Data().Set(JwtData, codekit.M{}.Set("CompanyID", companyid))
	return ctx
}

func InvokeAPI[M, R any](svc *kaos.Service, uriPath string, payload M, respond R, userid, coid string) (R, error) {
	if svc == nil {
		return respond, fmt.Errorf("missing: service, source: invokeAPI %s", uriPath)
	}

	sr := svc.GetRoute(uriPath)
	if sr == nil {
		return respond, fmt.Errorf("missing: route: %s", uriPath)
	}
	ctx := PrepareCtxData(kaos.NewContextFromService(svc, sr), userid, coid)
	e := svc.CallTo(uriPath, respond, ctx, payload)
	return respond, e
}
