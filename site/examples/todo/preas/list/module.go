package list

import (
	"github.com/prea-framework/prea/core"
	"github.com/prea-framework/prea/di"
)

var Module = &core.Module{
	Name: "list",
	Providers: []any{
		Repository{},
		di.Export(Service{}),
		Handler{},
	},
}
