package models

type Project struct {
	ID      int64
	Title   string
	OwnerID string `json:"-"`
}
