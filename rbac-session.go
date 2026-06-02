package rayiapp

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RbacSession struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	UserID            string
	Expired           time.Time
	Data              codekit.M
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *RbacSession) TableName() string {
	return "RbacSessions"
}

func (o *RbacSession) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RbacSession) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RbacSession) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RbacSession) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RbacSession) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RbacSession) PostSave(dbflex.IConnection) error {
	return nil
}
func (o RbacSession) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "RbacSession_UserID", Fields: []string{"UserID", "Expired"}},
		{Name: "RbacSession_Expired", Fields: []string{"Expired"}},
	}
}
