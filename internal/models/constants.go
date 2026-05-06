package models

const (
	RoleGuest     = "guest"
	RoleUser      = "user"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
)

const (
	TargetTypePost    = "post"
	TargetTypeComment = "comment"
)

const (
	ReportReasonIrrelevant = "irrelevant"
	ReportReasonObscene    = "obscene"
	ReportReasonIllegal    = "illegal"
	ReportReasonInsulting  = "insulting"
)

const (
	ReportStatusPending  = "pending"
	ReportStatusReviewed = "reviewed"
	ReportStatusRejected = "rejected"
)

const (
	NotificationTypeLike    = "like"
	NotificationTypeDislike = "dislike"
	NotificationTypeComment = "comment"
)