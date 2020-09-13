package neo

type SearchableEntity struct {
	ID       uint64 `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Type     string `db:"type" json:"type"`
	Image    string `db:"image" json:"image"`
	Priority int    `db:"priority" json:"-"`
}
