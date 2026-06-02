package rayiapp

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sebarcode/codekit"
)

type SessJWT struct {
	jwt.RegisteredClaims
	Data codekit.M
}

type JwtToSessOpts struct {
	Secret     string
	KaosCtx    *kaos.Context
	ValidateDb bool
}

func JwtToSession(token string, opts *JwtToSessOpts) (string, codekit.M, codekit.M, error) {
	var sessData codekit.M

	if opts == nil {
		return "", nil, sessData, errors.New("missing opts")
	}

	if opts.Secret == "" {
		return "", nil, sessData, errors.New("missing secret")
	}

	bc := new(SessJWT)
	m := codekit.M{}
	tkn, e := jwt.ParseWithClaims(token, bc, func(t *jwt.Token) (interface{}, error) {
		return []byte(opts.Secret), nil
	})
	if e != nil {
		return "", m, sessData, e
	}
	if !tkn.Valid {
		return bc.ID, bc.Data, sessData, errors.New("jwt token is invalid")
	}
	// check expiration using RegisteredClaims (ExpiresAt is *jwt.NumericDate)
	if bc.ExpiresAt == nil || bc.ExpiresAt.Time.Before(time.Now()) {
		return bc.ID, bc.Data, sessData, errors.New("jwt token is expired")
	}
	if opts.ValidateDb {
		if opts.KaosCtx == nil {
			return "", nil, sessData, errors.New("missing: kaos context")
		}
		db, _ := opts.KaosCtx.GetHub("rbac", "")
		if db == nil {
			return "", nil, nil, fmt.Errorf("missing: rbac db")
		}
		dbSess, _ := datahub.GetByID(db, new(RbacSession), bc.ID)
		SetCtxDataWithSessionInfo(opts.KaosCtx, dbSess, bc.Data)
		sessData = dbSess.Data
	}
	return bc.ID, bc.Data, sessData, nil
}

func SessionToJwt(signMethodName string, signSecret string, sess *RbacSession, lifeTime int, jwtData codekit.M) (string, error) {
	signMethod := jwt.GetSigningMethod(signMethodName)
	if signMethod == nil {
		return "", fmt.Errorf("invalid-sign-method: %s", signMethodName)
	}

	bc := new(SessJWT)
	// RegisteredClaims uses ID and ExpiresAt as a NumericDate
	bc.ID = sess.ID
	bc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(lifeTime) * time.Second))
	if jwtData == nil {
		jwtData = codekit.M{}
	}
	bc.Data = jwtData

	token := jwt.NewWithClaims(signMethod, bc)
	tokenString, err := token.SignedString([]byte(signSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetJwtToken(ctx *kaos.Context) string {
	tokenString, ok := ctx.Data().Get(CtxJwtToken, "").(string)
	if !ok {
		return ""
	}
	return tokenString
}

func GetJwtReferenceID(ctx *kaos.Context) string {
	referenceID, ok := ctx.Data().Get(CtxJwtReferenceID, "").(string)
	if !ok {
		return ""
	}
	return referenceID
}

func GetJwtClientData(ctx *kaos.Context) codekit.M {
	clientData, ok := ctx.Data().Get(CtxJwtClientData, "").(codekit.M)
	if !ok {
		return nil
	}
	return clientData
}

func GetJwtSessionData(ctx *kaos.Context) codekit.M {
	sessionData, ok := ctx.Data().Get(CtxJwtSessionData, "").(codekit.M)
	if !ok {
		return nil
	}
	return sessionData
}

func SetCtxDataWithSessionInfo(ctx *kaos.Context, dbSess *RbacSession, clientData codekit.M) {
	ctx.Data().Set(CtxJwtSessionID, dbSess.ID)
	ctx.Data().Set(CtxJwtReferenceID, dbSess.UserID)
	if dbSess.Data == nil {
		dbSess.Data = codekit.M{}
	}
	ctx.Data().Set(CtxJwtSessionData, dbSess.Data)
	ctx.Data().Set(CtxJwtClientData, clientData)
}
