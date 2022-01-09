package core

import (
	"time"

	"github.com/google/uuid"
)

// User represents user entity
type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Nickname  string
	Password  string
	Email     string
	Country   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserFilter struct {
	Country   string
	FirstName string
	LastName  string
	Nickname  string

	// Pagination
	PreviousPage string
	NextPage     string
	Limit        int
}
