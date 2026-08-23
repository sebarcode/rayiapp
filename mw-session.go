package rayiapp

import (
	"errors"
	"reflect"
	"slices"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
)

func storeToken(ctx *kaos.Context, needJwt bool) (string, error) {
	req := ctx.HttpRequest()
	if req == nil {
		return "", errors.New("missing http request")
	}
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		if needJwt {
			return "", errors.New("missing authorization header")
		} else {
			return "", nil
		}

	}
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		if needJwt {
			return "", errors.New("invalid authorization header format")
		}
	} else {
		tokenString := authHeader[len(bearerPrefix):]
		ctx.Data().Set("jwt_token", tokenString)
		return tokenString, nil
	}
	return "", nil
}

func MwStoreToken(needJwt bool) func(ctx *kaos.Context, _ any) (bool, error) {
	return func(ctx *kaos.Context, _ any) (bool, error) {
		_, err := storeToken(ctx, needJwt)
		return err == nil, err
	}
}

func MwValidateJWT(validateDb, continueIfInvalidJWT bool) func(ctx *kaos.Context, _ any) (bool, error) {
	return func(ctx *kaos.Context, _ any) (bool, error) {
		tokenString, _ := storeToken(ctx, true)
		if tokenString == "" {
			if !continueIfInvalidJWT {
				return false, errors.New("missing jwt token")
			}
		}
		ev, _ := ctx.DefaultEvent()
		if ev == nil {
			if !continueIfInvalidJWT {
				return false, errors.New("missing rbac event")
			}
		}
		request := ValidateJwtRequest{
			Token:       tokenString,
			GetSessData: validateDb,
		}
		response := ValidateJwtResponse{}
		err := ev.Publish("/rbac/validate-jwt", &request, &response, nil)
		if err != nil {
			if !continueIfInvalidJWT {
				return false, err
			}
		}
		SetCtxDataWithSessionInfo(ctx, &RbacSession{
			ID:     response.SessionID,
			UserID: response.UserID,
			Data:   response.SessionData,
		}, response.ClientData)
		return true, nil
	}
}

func MwCheckRole(role string) kaos.MWFunc {
	return func(ctx *kaos.Context, _ any) (bool, error) {
		roles := ctx.Data().Get("Roles", nil)
		if roles == nil {
			clientData := GetJwtClientData(ctx)
			if clientData != nil {
				roles = clientData.Get("Roles", nil)
			}
		}

		roleIds := []string{}
		switch values := roles.(type) {
		case []string:
			roleIds = values
		case []any:
			for _, value := range values {
				if roleID, ok := value.(string); ok {
					roleIds = append(roleIds, roleID)
				}
			}
		}
		if len(roleIds) == 0 {
			return false, errors.New("missing role in context")
		}
		if !slices.Contains(roleIds, role) {
			return false, errors.New("unauthorized_invalid_role_access")
		}
		return true, nil
	}
}

func MwCheckPolicy(policyid string, policyValue int) kaos.MWFunc {
	return func(ctx *kaos.Context, _ any) (bool, error) {
		policies, ok := ctx.Data().Get("Policies", map[string]int{}).(map[string]int)
		if !ok {
			return false, errors.New("unauthorized_invalid_policy_access")
		}
		if val, exists := policies[policyid]; exists {
			if val&policyValue == policyValue {
				return true, nil
			}
		}
		return false, errors.New("unauthorized_invalid_policy_access")
	}
}

func MwLimitTake(limit int) kaos.MWFunc {
	return func(ctx *kaos.Context, payload any) (bool, error) {
		if limit <= 0 {
			return false, errors.New("invalid limit value")
		}

		smPath := ctx.Data().Get("path", "").(string)
		if !(strings.HasSuffix(smPath, "/gets") || strings.HasSuffix(smPath, "/find")) {
			return true, nil
		}

		qp, ok := payload.(*dbflex.QueryParam)
		if !ok {
			return false, errors.New("payload is not a QueryParam")
		}
		take := qp.Take
		if take > limit || take == 0 {
			qp.Take = limit
		}
		val := reflect.ValueOf(payload)
		val.Elem().Set(reflect.ValueOf(qp).Elem())

		return true, nil
	}
}
