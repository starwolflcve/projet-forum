package models

import "time"

type ActivityItem struct {
	Type         string
	PostID       int
	CommentID    int
	PostTitle    string
	CommentBody  string
	ReactionType string
	CreatedAt    time.Time
}