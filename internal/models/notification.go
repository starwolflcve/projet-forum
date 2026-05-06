package models

import "time"

type Notification struct {
	ID        int
	UserID    int
	Type      string
	Message   string
	RelatedID int
	IsRead    bool
	CreatedAt time.Time
}