package models

import "time"

type Reaction struct {
    ID           int       `json:"id"`
    UserID       int       `json:"user_id"`
    TargetType   string    `json:"target_type"`
    TargetID     int       `json:"target_id"`
    ReactionType string    `json:"reaction_type"`
    CreatedAt    time.Time `json:"created_at"`
}
