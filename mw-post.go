package rayiapp

import (
	"errors"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/reflector"
	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
)

func MWPostExtractForeignField(model orm.DataModel, prefix, fieldName string, otherName ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		msRes := []codekit.M{}
		err := serde.Serde(ctx.Data().Get("FnResult", []codekit.M{}), &msRes)
		if err != nil {
			return false, err
		}

		db, _ := ctx.DefaultHub()
		if db == nil {
			return false, errors.New("missing database connection")
		}

		mrCat := NewMapRecordWithORM(db, model)

		for index, m := range msRes {
			catID := m.GetString(fieldName)
			cat, err := mrCat.Get(catID)
			if err != nil {
				continue
			}

			for _, other := range otherName {
				obj, _ := reflector.From(cat).Get(other)
				m.Set(prefix+other, obj)
			}
			msRes[index] = m
		}

		ctx.Data().Set("FnResult", msRes)

		return true, nil
	}
}
