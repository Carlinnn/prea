package main

import (
	"github.com/prea-framework/prea"

	"github.com/prea-framework/prea/example/preas"
)

func main() {
	prea.Run(prea.Config{
		Modules:     preas.Modules,
		Middlewares: preas.Middlewares,
		Security:    preas.Security,
		Logging:     preas.Logging,
	})
}
