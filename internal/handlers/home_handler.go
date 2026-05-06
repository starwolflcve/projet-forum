package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "projet-forum/internal/database"
    "projet-forum/internal/models"
)

var repository *database.PostRepository

func SetRepository(repo *database.PostRepository) {
    repository = repo
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    filter := models.PostFilter{
        CategorySlug:  r.URL.Query().Get("category"),
        AuthorName:    r.URL.Query().Get("author"),
        SearchQuery:   r.URL.Query().Get("q"),
        SortBy:        r.URL.Query().Get("sort_by"),
    }

    if authorID, err := strconv.Atoi(r.URL.Query().Get("author_id")); err == nil {
        filter.AuthorID = authorID
    }
    if likedBy, err := strconv.Atoi(r.URL.Query().Get("liked_by")); err == nil {
        filter.LikedByUserID = likedBy
    }

    posts, err := repository.ListPosts(filter)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    categories, err := repository.ListCategories()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(categories)
}

func parseIDQuery(r *http.Request, key string) (int, error) {
    return strconv.Atoi(r.URL.Query().Get(key))
}

func parseUserID(r *http.Request) (int, error) {
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        userID = r.URL.Query().Get("user_id")
    }
    return strconv.Atoi(userID)
}

func writeJSON(w http.ResponseWriter, value interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(value)
}
