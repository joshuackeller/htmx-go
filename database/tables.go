package database

import (
	"time"
)

type Todo struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

func (Todo) TableName() string {
	return "todo"
}
