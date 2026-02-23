package types

type Entry struct {
	Title string
	Url   string
	Tags  []string
}

func (e *Entry) GetTitle() string {
	if e.Title == "" {
		return e.Url
	}

	return e.Title
}
