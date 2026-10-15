package task

// #region model
// Task is one thing to do. It may belong to a list.
type Task struct {
	ID     string  `json:"id" pk:"" generated:"uuid"`
	Title  string  `json:"title" size:"200"`
	Done   bool    `json:"done" default:"false"`
	ListID *string `json:"listId" fk:"list" ondelete:"cascade" index:""`
}

func (Task) Table() string {
	return "task"
}

// #endregion model
