package rayiapp

import (
	"fmt"
	"strings"

	"errors"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/dbmod"
)

func NewDBModFromContext() kaos.Mod {
	dbm := dbmod.New()
	dbm.SetHubFn(GetTenantDBFromContext)
	return dbm
}

func AddTenantDB(appConfig *AppConfig, hubs map[string]*datahub.Hub, tenantID string) map[string]*datahub.Hub {
	tenantConn := appConfig.Connections["tenant"]
	tenantConnStr := fmt.Sprintf(tenantConn.Txt, tenantID)
	hubs["tenant_"+tenantID] = datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(tenantConnStr, true), true, tenantConn.PoolSize)
	return hubs
}

func GetTenantDB(ctx *kaos.Context, name string) *datahub.Hub {
	h, _ := ctx.GetHub(name, "tenant")
	return h
}

func GetTenantDBFromContext(ctx *kaos.Context) *datahub.Hub {
	jwtdata := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
	tenantID := jwtdata.GetString("TenantID")
	if tenantID == "" {
		tenantID = "Demo"
	}
	h, _ := ctx.GetHub(tenantID, "tenant")
	return h
}

// DBError  check  if database error
func DBError(e error) bool {
	return e != nil && strings.Contains(e.Error(), "EOF")
}

// Tx datahub transaction safe wrapper
func Tx(h *datahub.Hub, allowNonTx bool, fn func(tx *datahub.Hub) error) (e error) {
	var tx *datahub.Hub

	if h.IsTx() {
		tx = h
	} else {
		tx, e = h.BeginTx()
		if e == nil && tx.IsTx() {
			// do nothing
		} else if e == nil && !tx.IsTx() && !allowNonTx {
			return errors.New("missing: tx")
		} else if e != nil && !allowNonTx {
			return e
		} else {
			tx = h
		}
	}

	defer func() {
		var eTx error

		if r := recover(); r != nil {
			e = fmt.Errorf("db tx panic: %s", r)
			h.Log().Errorf("%s", e.Error())
		}

		if !tx.IsTx() {
			return
		}

		if e != nil {
			eTx = tx.Rollback()
			if eTx != nil {
				e = fmt.Errorf("rollback db: %s", eTx.Error())
			}
			return
		}

		eTx = tx.Commit()
		if eTx != nil {
			e = fmt.Errorf("commit db: %s", eTx.Error())
		}
	}()

	if e = fn(tx); e != nil {
		return e
	}

	return nil
}
