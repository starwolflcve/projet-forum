package database

import (
	"database/sql"
	"time"

	"forum/internal/models"
)

func CreateReportsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		reporter_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		reason TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		reviewed_by INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		reviewed_at DATETIME,
		FOREIGN KEY (reporter_id) REFERENCES users(id),
		FOREIGN KEY (reviewed_by) REFERENCES users(id)
	);`

	_, err := db.Exec(query)
	return err
}

func CreateNotificationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		message TEXT NOT NULL,
		related_id INTEGER,
		is_read INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	_, err := db.Exec(query)
	return err
}

func CreateP3Indexes(db *sql.DB) error {
	queries := []string{
		`CREATE INDEX IF NOT EXISTS idx_reports_reporter_id ON reports(reporter_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reports_target ON reports(target_type, target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func InitP3Tables(db *sql.DB) error {
	if err := CreateReportsTable(db); err != nil {
		return err
	}

	if err := CreateNotificationsTable(db); err != nil {
		return err
	}

	if err := CreateP3Indexes(db); err != nil {
		return err
	}

	return nil
}

func CreateReport(db *sql.DB, report *models.Report) error {
	query := `
	INSERT INTO reports (reporter_id, target_type, target_id, reason, status, reviewed_by, created_at, reviewed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(
		query,
		report.ReporterID,
		report.TargetType,
		report.TargetID,
		report.Reason,
		report.Status,
		report.ReviewedBy,
		report.CreatedAt,
		report.ReviewedAt,
	)

	return err
}

func ListPendingReports(db *sql.DB) ([]models.Report, error) {
	query := `
	SELECT id, reporter_id, target_type, target_id, reason, status, reviewed_by, created_at, reviewed_at
	FROM reports
	WHERE status = ?
	ORDER BY created_at DESC
	`

	rows, err := db.Query(query, models.ReportStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.Report

	for rows.Next() {
		var report models.Report
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&report.TargetType,
			&report.TargetID,
			&report.Reason,
			&report.Status,
			&reviewedBy,
			&report.CreatedAt,
			&reviewedAt,
		)
		if err != nil {
			return nil, err
		}

		if reviewedBy.Valid {
			report.ReviewedBy = int(reviewedBy.Int64)
		}

		if reviewedAt.Valid {
			t := reviewedAt.Time
			report.ReviewedAt = &t
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}

func UpdateReportStatus(db *sql.DB, reportID int, status string, reviewedBy int) error {
	query := `
	UPDATE reports
	SET status = ?, reviewed_by = ?, reviewed_at = ?
	WHERE id = ?
	`

	_, err := db.Exec(query, status, reviewedBy, time.Now(), reportID)
	return err
}

func CreateNotification(db *sql.DB, notification *models.Notification) error {
	query := `
	INSERT INTO notifications (user_id, type, message, related_id, is_read, created_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(
		query,
		notification.UserID,
		notification.Type,
		notification.Message,
		notification.RelatedID,
		notification.IsRead,
		notification.CreatedAt,
	)

	return err
}

func ListNotificationsByUserID(db *sql.DB, userID int) ([]models.Notification, error) {
	query := `
	SELECT id, user_id, type, message, related_id, is_read, created_at
	FROM notifications
	WHERE user_id = ?
	ORDER BY created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification

	for rows.Next() {
		var notification models.Notification
		var isRead int

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Message,
			&notification.RelatedID,
			&isRead,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		notification.IsRead = isRead == 1
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func MarkNotificationAsRead(db *sql.DB, notificationID int, userID int) error {
	query := `
	UPDATE notifications
	SET is_read = 1
	WHERE id = ? AND user_id = ?
	`

	_, err := db.Exec(query, notificationID, userID)
	return err
}

func ListUserCreatedPosts(db *sql.DB, userID int) ([]models.ActivityItem, error) {
	query := `
	SELECT id, title, created_at
	FROM posts
	WHERE user_id = ?
	ORDER BY created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ActivityItem

	for rows.Next() {
		var item models.ActivityItem

		err := rows.Scan(
			&item.PostID,
			&item.PostTitle,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.Type = "post"
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func ListUserReactions(db *sql.DB, userID int) ([]models.ActivityItem, error) {
	query := `
	SELECT r.target_id, p.title, r.reaction_type, r.created_at
	FROM reactions r
	JOIN posts p ON p.id = r.target_id
	WHERE r.user_id = ? AND r.target_type = 'post'
	ORDER BY r.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ActivityItem

	for rows.Next() {
		var item models.ActivityItem

		err := rows.Scan(
			&item.PostID,
			&item.PostTitle,
			&item.ReactionType,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.Type = "reaction"
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func ListUserCommentsActivity(db *sql.DB, userID int) ([]models.ActivityItem, error) {
	query := `
	SELECT c.id, c.post_id, p.title, c.content, c.created_at
	FROM comments c
	JOIN posts p ON p.id = c.post_id
	WHERE c.user_id = ?
	ORDER BY c.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ActivityItem

	for rows.Next() {
		var item models.ActivityItem

		err := rows.Scan(
			&item.CommentID,
			&item.PostID,
			&item.PostTitle,
			&item.CommentBody,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.Type = "comment"
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func ListActivityByUserID(db *sql.DB, userID int) (map[string][]models.ActivityItem, error) {
	posts, err := ListUserCreatedPosts(db, userID)
	if err != nil {
		return nil, err
	}

	reactions, err := ListUserReactions(db, userID)
	if err != nil {
		return nil, err
	}

	comments, err := ListUserCommentsActivity(db, userID)
	if err != nil {
		return nil, err
	}

	return map[string][]models.ActivityItem{
		"posts":     posts,
		"reactions": reactions,
		"comments":  comments,
	}, nil
}

func DeletePostByID(db *sql.DB, postID int) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := db.Exec(query, postID)
	return err
}

func DeleteCommentByID(db *sql.DB, commentID int) error {
	query := `DELETE FROM comments WHERE id = ?`
	_, err := db.Exec(query, commentID)
	return err
}

func CreateCategory(db *sql.DB, name string) error {
	query := `
	INSERT INTO categories (name)
	VALUES (?)
	`
	_, err := db.Exec(query, name)
	return err
}

func DeleteCategoryByID(db *sql.DB, categoryID int) error {
	query := `DELETE FROM categories WHERE id = ?`
	_, err := db.Exec(query, categoryID)
	return err
}

func ListCategories(db *sql.DB) ([]models.Category, error) {
	query := `
	SELECT id, name
	FROM categories
	ORDER BY name ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		var category models.Category

		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func UpdateUserRole(db *sql.DB, userID int, role string) error {
	query := `
	UPDATE users
	SET role = ?
	WHERE id = ?
	`
	_, err := db.Exec(query, role, userID)
	return err
}

func ListUsers(db *sql.DB) ([]models.User, error) {
	query := `
	SELECT id, username, email, role
	FROM users
	ORDER BY username ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func DeleteOwnPostByID(db *sql.DB, postID int, userID int) error {
	query := `
	DELETE FROM posts
	WHERE id = ? AND user_id = ?
	`
	_, err := db.Exec(query, postID, userID)
	return err
}

func DeleteOwnCommentByID(db *sql.DB, commentID int, userID int) error {
	query := `
	DELETE FROM comments
	WHERE id = ? AND user_id = ?
	`
	_, err := db.Exec(query, commentID, userID)
	return err
}