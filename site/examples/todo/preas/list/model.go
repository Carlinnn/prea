package list

// #region model
// List groups tasks: "Groceries", "Work"...
type List struct {
	ID   string `json:"id" pk:"" generated:"uuid"`
	Name string `json:"name" size:"80" unique:""`
}

func (List) Table() string {
	return "list"
}

// #endregion model
