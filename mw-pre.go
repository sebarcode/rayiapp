package rayiapp

import (
	"errors"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/reflector"
)

// MwInjectTenant is a pre-middleware that injects TenantID into the payload
// on write operations (insert, update, save). It should be registered after
// the JWT middleware so that the tenant ID is available in the context.
func MwInjectTenant(userFieldName string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		svcPath := ctx.Data().Get("path", "").(string)
		if strings.HasSuffix(svcPath, "insert") || strings.HasSuffix(svcPath, "update") || strings.HasPrefix(svcPath, "save") {
			tenantID := GetTenantID(ctx)
			r := reflector.From(payload)
			r.Set("TenantID", tenantID)
			if userFieldName != "" {
				userID := GetJwtReferenceID(ctx)
				r.Set(userFieldName, userID)
			}
			err := r.Flush()
			if err != nil {
				return false, ctx.Log().Error2("fail to set TenantID", "fail to set TenantID object %v %s", payload, err)
			}
		}
		return true, nil
	}
}

// MwFilterTenant is a pre-middleware that injects a TenantID filter into
// read operations (gets and find). It should be registered after the JWT
// middleware so that the tenant ID is available in the context.
func MwFilterTenant() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		path := ctx.Data().Get("path", "").(string)
		if !(strings.HasSuffix(path, "/gets") || strings.HasSuffix(path, "/find")) {
			return true, nil
		}

		tenantID := GetTenantID(ctx)
		qp, ok := payload.(*dbflex.QueryParam)
		if !ok {
			if payload != nil {
				return false, errors.New("payload is not a QueryParam")
			}
		}
		if payload == nil {
			qp = dbflex.NewQueryParam()
		}
		qp.MergeWhere(false, dbflex.Eq("TenantID", tenantID))

		return true, nil
	}
}
