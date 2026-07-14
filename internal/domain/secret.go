package domain

import "time"

type Secret struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
