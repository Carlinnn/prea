package task

import "github.com/prea-framework/prea/database"

// Repository reads and writes the task table.
type Repository struct {
	database.BaseRepository[Task]
}
