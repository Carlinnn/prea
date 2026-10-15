package list

import "github.com/prea-framework/prea/database"

// Repository reads and writes the list table.
type Repository struct {
	database.BaseRepository[List]
}
