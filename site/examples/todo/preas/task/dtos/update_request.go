package dtos

// #region update
// UpdateRequest is the body of PATCH /task/:id. Only the fields you send
// change: nil keeps the column as it is.
type UpdateRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

// #endregion update
