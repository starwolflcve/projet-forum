package models

import "time"

type Post struct {
    ID               int       `json:"id"`
    UserID           int       `json:"user_id"`
    AuthorName       string    `json:"author_name"`
    Title            string    `json:"title"`
    Content          string    `json:"content"`
    ImagePath        string    `json:"image_path"`
    Visibility       string    `json:"visibility"`
    ModerationStatus string    `json:"moderation_status"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

type PostFilter struct {
    CategorySlug  string `json:"category_slug"`
    AuthorID      int    `json:"author_id"`
    AuthorName    string `json:"author_name"`
    LikedByUserID int    `json:"liked_by_user_id"`
    SearchQuery   string `json:"search_query"`
    SortBy        string `json:"sort_by"`
}
