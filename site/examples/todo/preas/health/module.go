package health

import "github.com/prea-framework/prea/core"

var Module = &core.Module{
	Name:      "health",
	Providers: []any{Handler{}},
}
