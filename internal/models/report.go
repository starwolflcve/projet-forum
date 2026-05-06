package models

import "time"

type Report struct {
	ID         int
	ReporterID int
	TargetType string
	TargetID   int
	Reason     string
	Status     string
	ReviewedBy int
	CreatedAt  time.Time
	ReviewedAt *time.Time
}