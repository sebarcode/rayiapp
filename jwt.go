package rayiapp

import (
	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

func JwtClientData(ctx *kaos.Context) codekit.M {
	m, ok := ctx.Data().Get(CtxJwtClientData, codekit.M{}).(codekit.M)
	if !ok {
		return codekit.M{}
	}
	return m
}

func JwtSessionData(ctx *kaos.Context) codekit.M {
	m, ok := ctx.Data().Get(CtxJwtSessionData, codekit.M{}).(codekit.M)
	if !ok {
		return codekit.M{}
	}
	return m
}

func JwtToken(ctx *kaos.Context) string {
	m, ok := ctx.Data().Get(CtxJwtToken, "").(string)
	if !ok {
		return ""
	}
	return m
}

func JwtUserID(ctx *kaos.Context) string {
	m, ok := ctx.Data().Get(CtxJWTReferenceID, "").(string)
	if !ok {
		return ""
	}
	return m
}
