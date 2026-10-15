package list

import (
	"context"
	"net/http"
	"strings"

	"github.com/prea-framework/prea/core"

	"todo/preas/list/dtos"
)

// Service holds the rules of the list preá.
type Service struct {
	Repo *Repository `inject:""`
}

func (s *Service) List(ctx context.Context) ([]List, error) {
	return s.Repo.GetAll(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (List, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req dtos.CreateRequest) (List, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return List{}, core.NewError(http.StatusBadRequest, "name can't be blank")
	}
	return s.Repo.Create(ctx, req)
}

func (s *Service) Update(ctx context.Context, id string, req dtos.UpdateRequest) (List, error) {
	req.Name = strings.TrimSpace(req.Name)
	return s.Repo.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}
