package task

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/prea-framework/prea/core"
	"github.com/prea-framework/prea/database"

	"todo/preas/list"
	"todo/preas/task/dtos"
)

// #region service
// Service holds the rules of a task.
type Service struct {
	Repo  *Repository   `inject:""`
	Lists *list.Service `inject:""` // exported by the list preá
}

// #endregion service

// #region list
// List returns every task, or only the ones of a list.
func (s *Service) List(ctx context.Context, listID string) ([]Task, error) {
	if listID == "" {
		return s.Repo.GetAll(ctx)
	}
	return s.Repo.Where(ctx, database.Args{"list_id": listID})
}

// #endregion list

func (s *Service) Get(ctx context.Context, id string) (Task, error) {
	return s.Repo.GetByID(ctx, id)
}

// #region create
func (s *Service) Create(ctx context.Context, req dtos.CreateRequest) (Task, error) {
	req.Title = strings.TrimSpace(req.Title)
	if err := checkTitle(req.Title); err != nil {
		return Task{}, err
	}
	if req.ListID != nil {
		if err := s.checkList(ctx, *req.ListID); err != nil {
			return Task{}, err
		}
	}
	return s.Repo.Create(ctx, req)
}

// #endregion create

func (s *Service) Update(ctx context.Context, id string, req dtos.UpdateRequest) (Task, error) {
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if err := checkTitle(title); err != nil {
			return Task{}, err
		}
		req.Title = &title
	}
	return s.Repo.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}

// #region rules
const maxTitle = 200

func checkTitle(title string) error {
	switch {
	case title == "":
		return core.NewError(http.StatusBadRequest, "title can't be blank")
	case utf8.RuneCountInString(title) > maxTitle:
		return core.NewError(http.StatusBadRequest, "title is too long: 200 characters at most")
	}
	return nil
}

// checkList turns a missing list into a clear 422, instead of the
// database's foreign key error.
func (s *Service) checkList(ctx context.Context, id string) error {
	_, err := s.Lists.Get(ctx, id)
	if errors.Is(err, database.ErrNotFound) {
		return core.NewError(http.StatusUnprocessableEntity, "list not found")
	}
	return err
}

// #endregion rules
