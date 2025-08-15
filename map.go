package rayiapp

import (
	"errors"
	"reflect"

	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/datahub"
)

type MapRecord[T any] struct {
	records map[string]T
	getFn   func(key string) (T, error)
}

func NewMapRecord[T any](getFn func(key string) (T, error)) *MapRecord[T] {
	mr := new(MapRecord[T])
	mr.records = map[string]T{}
	mr.getFn = getFn
	return mr
}

func NewMapRecordWithORM[T orm.DataModel](db *datahub.Hub, model T) *MapRecord[T] {
	rt := reflect.TypeOf(model).Elem()

	mr := new(MapRecord[T])
	mr.records = map[string]T{}
	mr.getFn = func(key string) (T, error) {
		model := reflect.New(rt).Interface().(T)
		return datahub.GetByID(db, model, key)
	}
	return mr
}

func (mr *MapRecord[T]) Get(key string) (T, error) {
	var err error
	if mr.records == nil {
		mr.records = map[string]T{}
	}
	record, ok := mr.records[key]
	if !ok {
		if mr.getFn == nil {
			return record, errors.New("mapRecord getFn is nil")
		}
		if record, err = mr.getFn(key); err != nil {
			return record, err
		}
		mr.records[key] = record
	}
	return record, nil
}

func (mr *MapRecord[T]) Keys() []string {
	keys := make([]string, len(mr.records))
	idx := 0
	for k := range mr.records {
		keys[idx] = k
		idx++
	}
	return keys
}

func (mr *MapRecord[T]) Records() []T {
	records := make([]T, len(mr.records))
	idx := 0
	for _, v := range mr.records {
		records[idx] = v
		idx++
	}
	return records
}
