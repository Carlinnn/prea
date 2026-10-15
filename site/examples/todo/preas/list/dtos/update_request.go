package dtos

// UpdateRequest is the body of PATCH /list/:id. Zero fields are not
// updated; a pointer field can set a zero value.
type UpdateRequest struct {
	Name string `json:"name"`
}
