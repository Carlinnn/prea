package di_test

import (
	"fmt"

	"github.com/prea-framework/prea/di"
)

type Repository struct{}

func (*Repository) Find() string { return "a task" }

type Service struct {
	Repo *Repository `inject:""`
}

// A module declares its components as values; the container creates one of
// each and fills the inject fields.
func Example() {
	module := &di.Module{
		Name:      "task",
		Providers: []any{Repository{}, Service{}},
	}

	c := di.New()
	if err := c.Load(module); err != nil {
		fmt.Println(err)
		return
	}
	for _, p := range c.Providers(module) {
		if s, ok := p.(*Service); ok {
			fmt.Println(s.Repo.Find())
		}
	}
	// Output: a task
}

// A missing dependency is an error when the module loads, with the module,
// the struct and the field.
func Example_missing() {
	module := &di.Module{Name: "task", Providers: []any{Service{}}}

	err := di.New().Load(module)
	fmt.Println(err)
	// Output: di: module task: Service.Repo: no provider for *di_test.Repository
}
