package rayiapp

import (
	"context"
	"testing"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

func TestCopyContextDataToPublishOptionsForwardsTenantIDAsString(t *testing.T) {
	ctx := kaos.NewContext(context.Background(), nil, nil, nil, kaos.NewSharedData())
	ctx.Data().Set(CtxJwtClientData, codekit.M{"TenantID": "tenant-1"})

	opts := CopyContextDataToPublishOptions(ctx, &kaos.PublishOpts{}, CtxJwtClientData)
	if got := opts.Headers.GetString("TenantID"); got != "tenant-1" {
		t.Fatalf("expected TenantID header tenant-1, got %q", got)
	}
	if _, ok := opts.Headers[CtxJwtClientData].(codekit.M); !ok {
		t.Fatal("expected requested context data to remain available in publish options")
	}
}
