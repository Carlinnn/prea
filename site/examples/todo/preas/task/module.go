package task

import (
	"github.com/prea-framework/prea/core"
	"github.com/prea-framework/prea/di"

	"todo/preas/list"
)

// #region module
var Module = &core.Module{
	Name:    "task",
	Imports: []*core.Module{list.Module}, // to use list.Service
	Providers: []any{
		Repository{},
		di.Export(Service{}),
		Handler{},
	},
}

// #endregion module
