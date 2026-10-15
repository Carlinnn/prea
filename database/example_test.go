package database_test

import (
	"context"

	"github.com/prea-framework/prea/database"
)

type Task struct {
	ID    string `json:"id" pk:"" generated:"uuid"`
	Title string `json:"title" size:"200"`
	Done  bool   `json:"done" default:"false"`
}

func (Task) Table() string { return "task" }

// Repository gets every BaseRepository method for Task, and adds its own
// queries with Prepare.
type Repository struct {
	database.BaseRepository[Task]
}

func (r *Repository) Search(ctx context.Context, text string) ([]Task, error) {
	return r.Prepare("select * from task where title ilike @q order by title").
		Bind("q", "%"+text+"%").
		All(ctx)
}

// A repository embeds BaseRepository; the container injects its pool.
func ExampleBaseRepository() {
	var r Repository // declared as Repository{} in a module's Providers
	ctx := context.Background()

	task, _ := r.Create(ctx, struct{ Title string }{"Buy coffee"})
	_, _ = r.Update(ctx, task.ID, database.Args{"done": true})
	_, _ = r.Where(ctx, database.Args{"done": false})
	_, _ = r.Search(ctx, "coffee")
}

type Service struct {
	database.Transactional
	Tasks *Repository `inject:""`
}

// Every query that runs with the ctx of the function joins the transaction.
func ExampleTransactional_Transaction() {
	var s Service
	_ = s.Transaction(context.Background(), func(ctx context.Context) error {
		task, err := s.Tasks.Create(ctx, struct{ Title string }{"Buy coffee"})
		if err != nil {
			return err // rolls back
		}
		_, err = s.Tasks.Update(ctx, task.ID, database.Args{"done": true})
		return err // nil commits
	})
}
