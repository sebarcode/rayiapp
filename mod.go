package rayiapp

import (
	"fmt"

	"git.kanosolution.net/kano/kaos"
)

var (
	mapIntegrations = map[string][]string{}
)

func PostMWIntegration(eventHubName string) func(ctx *kaos.Context, payload interface{}) (bool, error) {
	if eventHubName == "" {
		eventHubName = "integration"
	}

	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		apiPath := ctx.Data().Get("path", "").(string)
		if apiPath == "" {
			return true, nil
		}
		evi := ctx.EventHubs()[eventHubName]
		if evi == nil {
			return true, nil
		}
		go func() {
			subjects, has := mapIntegrations[apiPath]
			if has {
				for _, subject := range subjects {
					fnRes := ctx.Data().Get("FnResult", nil)
					go evi.Publish(
						subject,
						fnRes, nil, &kaos.PublishOpts{
							Headers: ctx.Data().Data(),
						})
				}
			}
		}()
		return true, nil
	}
}

func RegisterIntegration(apiPath string, subject string) error {
	subjects, has := mapIntegrations[apiPath]
	if !has {
		subjects = []string{}
	}

	for _, exSubject := range subjects {
		if exSubject == subject {
			return fmt.Errorf("integration already exist, path: %s, integration: %s", apiPath, subject)
		}
	}

	subjects = append(subjects, subject)
	mapIntegrations[apiPath] = subjects

	return nil
}
