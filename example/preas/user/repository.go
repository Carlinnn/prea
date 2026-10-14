package user

import (
	"context"

	"github.com/prea-framework/prea/database"
)

type Repository struct {
	database.BaseRepository[User]
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (User, error) {
	return r.Prepare("select * from users where username = @username").
		Bind("username", username).
		One(ctx)
}
