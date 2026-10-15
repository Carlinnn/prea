package dtos

// CreateRequest is the body of POST /list.
type CreateRequest struct {
	Name string `json:"name" binding:"required"`
}
