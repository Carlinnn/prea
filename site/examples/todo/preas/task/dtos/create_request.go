package dtos

// #region create
// CreateRequest is the body of POST /task.
type CreateRequest struct {
	Title  string  `json:"title" binding:"required"`
	ListID *string `json:"listId"`
}

// #endregion create
