package rayiapp

import (
	"errors"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
)

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
			return false, errors.New("payload is not a QueryParam")
		}
		if qp == nil {
			qp = dbflex.NewQueryParam()
		}
		qp.MergeWhere(false, dbflex.Eq("TenantID", tenantID))

		return true, nil
	}
}
