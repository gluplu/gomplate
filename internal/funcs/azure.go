package funcs

import (
	"context"
	"sync"

	"github.com/hairyhenderson/gomplate/v4/azure"
)

// CreateGCPFuncs -
func CreateAzureFuncs(ctx context.Context) map[string]interface{} {
	ns := &AzureFuncs{
		ctx:     ctx,
		gcpopts: azure.GetClientOptions(),
	}
	return map[string]interface{}{
		"azure": func() interface{} { return ns },
	}
}

// GcpFuncs -
type AzureFuncs struct {
	ctx context.Context

	meta    *azure.MetaClient
	gcpopts azure.ClientOptions
}

// Meta -
func (a *AzureFuncs) Meta(key string, def ...string) (string, error) {
	a.meta = sync.OnceValue[*azure.MetaClient](func() *azure.MetaClient {
		return azure.NewMetaClient(a.ctx, a.gcpopts)
	})()

	return a.meta.Meta(key, def...)
}
