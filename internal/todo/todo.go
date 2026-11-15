// Package todo defines the todo domain model.
package todo

import "time"

// Todo is a single todo item.
type Todo struct {
	ID          int64
	Title       string
	Description *string // nil when not set
	Completed   bool
	DueDate     *time.Time // nil when not set; only the date part is used
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
