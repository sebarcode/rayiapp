package rayiapp

import (
	"errors"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/reflector"
	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
)

func MWPostFindExtractForeignField(model orm.DataModel, prefix, fieldName string, otherName ...string) kaos.MWFunc {
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

func MWPostFindExtractForeignFieldArray(model orm.DataModel, prefix, fieldName string, otherName ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		/*
			msRes := []codekit.M{}
			resBytes := codekit.Jsonify(ctx.Data().Get("FnResult", "[]"))
			err := codekit.Unjson(resBytes, &msRes)
		*/
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
			catIDs := m.Get(fieldName, []string{}).([]string)
			mapOthers := map[string][]interface{}{}
			for _, other := range otherName {
				mapOthers[other] = []interface{}{}
			}
			for _, catID := range catIDs {
				cat, err := mrCat.Get(catID)
				if err != nil {
					continue
				}

				for _, other := range otherName {
					obj, _ := reflector.From(cat).Get(other)
					mapOthers[other] = append(mapOthers[other], obj)
				}
			}
			for other, arrObj := range mapOthers {
				m.Set(prefix+other, arrObj)
			}
			msRes[index] = m
		}

		ctx.Data().Set("FnResult", msRes)

		return true, nil
	}
}

func MWPostGetsExtractForeignField(model orm.DataModel, prefix, fieldName string, otherName ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		msRes := []codekit.M{}
		mOrigRes := ctx.Data().Get("FnResult", codekit.M{}).(codekit.M)
		err := serde.Serde(mOrigRes.Get("data"), &msRes)
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

		mOrigRes.Set("data", msRes)
		ctx.Data().Set("FnResult", mOrigRes)

		return true, nil
	}
}

func MWPostGetsExtractForeignFieldArray(model orm.DataModel, prefix, fieldName string, otherName ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		msRes := []codekit.M{}
		mOrigRes := ctx.Data().Get("FnResult", codekit.M{}).(codekit.M)
		err := serde.Serde(mOrigRes.Get("data"), &msRes)
		if err != nil {
			return false, err
		}

		db, _ := ctx.DefaultHub()
		if db == nil {
			return false, errors.New("missing database connection")
		}

		mrCat := NewMapRecordWithORM(db, model)

		for index, m := range msRes {
			catIDs := m.Get(fieldName, []string{}).([]string)
			mapOthers := map[string][]interface{}{}
			for _, other := range otherName {
				mapOthers[other] = []interface{}{}
			}
			for _, catID := range catIDs {
				cat, err := mrCat.Get(catID)
				if err != nil {
					continue
				}

				for _, other := range otherName {
					obj, _ := reflector.From(cat).Get(other)
					mapOthers[other] = append(mapOthers[other], obj)
				}
			}
			for other, arrObj := range mapOthers {
				m.Set(prefix+other, arrObj)
			}
			msRes[index] = m
		}

		mOrigRes.Set("data", msRes)
		ctx.Data().Set("FnResult", mOrigRes)

		return true, nil
	}
}
