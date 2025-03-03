package rayiapp

import (
	"git.kanosolution.net/kano/kaos"
)

func PostMWIntegration(ctx *kaos.Context, payload interface{}) (bool, error) {
	apiPath := ctx.Data().Get("path", "").(string)
	if apiPath == "" {
		return true, nil
	}
	evi := ctx.EventHubs()["integration"]
	if evi == nil {
		return true, nil
	}
	go func() {
		fnRes := ctx.Data().Get("FnResult", nil)
		evi.Publish(apiPath, fnRes, nil, &kaos.PublishOpts{
			Headers: ctx.Data().Data(),
		})
	}()
	return true, nil
}
