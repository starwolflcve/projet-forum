package handlers

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "projet-forum/internal/models"
)

type createPostRequest struct {
    Title            string `json:"title"`
    Content          string `json:"content"`
    ImagePath        string `json:"image_path"`
    Visibility       string `json:"visibility"`
    ModerationStatus string `json:"moderation_status"`
    AuthorName       string `json:"author_name"`
    CategoryIDs      []int  `json:"category_ids"`
}

type reactRequest struct {
    TargetID     int    `json:"target_id"`
    ReactionType string `json:"reaction_type"`
}

func PostDetailHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    postID, err := parseIDQuery(r, "id")
    if err != nil || postID <= 0 {
        http.Error(w, "missing or invalid post id", http.StatusBadRequest)
        return
    }

    post, err := repository.GetPostByID(postID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if post == nil {
        http.Error(w, "post not found", http.StatusNotFound)
        return
    }

    comments, err := repository.ListCommentsByPostID(postID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    likes, _ := repository.CountPostLikes(postID)
    dislikes, _ := repository.CountPostDislikes(postID)

    response := map[string]interface{}{
        "post":      post,
        "comments":  comments,
        "likes":     likes,
        "dislikes":  dislikes,
    }
    writeJSON(w, response)
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req createPostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON payload", http.StatusBadRequest)
        return
    }

    userID, err := parseUserID(r)
    if err != nil || userID <= 0 {
        http.Error(w, "missing or invalid user_id", http.StatusBadRequest)
        return
    }

    authorName := strings.TrimSpace(req.AuthorName)
    if authorName == "" {
        authorName = fmt.Sprintf("User%d", userID)
    }

    post := &models.Post{
        UserID:           userID,
        AuthorName:       authorName,
        Title:            req.Title,
        Content:          req.Content,
        ImagePath:        req.ImagePath,
        Visibility:       req.Visibility,
        ModerationStatus: req.ModerationStatus,
    }

    postID, err := repository.CreatePost(post, req.CategoryIDs)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]interface{}{"id": postID})
}

func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    if r.Method != http.MethodPut && r.Method != http.MethodPatch {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    postID, err := parseIDQuery(r, "id")
    if err != nil || postID <= 0 {
        http.Error(w, "missing or invalid post id", http.StatusBadRequest)
        return
    }

    var req createPostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON payload", http.StatusBadRequest)
        return
    }

    existing, err := repository.GetPostByID(postID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if existing == nil {
        http.Error(w, "post not found", http.StatusNotFound)
        return
    }

    existing.Title = req.Title
    existing.Content = req.Content
    existing.ImagePath = req.ImagePath
    existing.Visibility = req.Visibility
    existing.ModerationStatus = req.ModerationStatus
    if req.AuthorName != "" {
        existing.AuthorName = strings.TrimSpace(req.AuthorName)
    }

    if err := repository.UpdatePost(existing); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if req.CategoryIDs != nil {
        if err := repository.AttachCategoriesToPost(postID, req.CategoryIDs); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    writeJSON(w, map[string]string{"status": "updated"})
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    if r.Method != http.MethodDelete {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    postID, err := parseIDQuery(r, "id")
    if err != nil || postID <= 0 {
        http.Error(w, "missing or invalid post id", http.StatusBadRequest)
        return
    }

    if err := repository.DeletePost(postID); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]string{"status": "deleted"})
}

func ReactToPostHandler(w http.ResponseWriter, r *http.Request) {
    reactHandler(w, r, "post")
}

func reactHandler(w http.ResponseWriter, r *http.Request, targetType string) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req reactRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON payload", http.StatusBadRequest)
        return
    }

    if req.TargetID <= 0 || req.ReactionType == "" {
        http.Error(w, "missing reaction target or type", http.StatusBadRequest)
        return
    }

    userID, err := parseUserID(r)
    if err != nil || userID <= 0 {
        http.Error(w, "missing or invalid user_id", http.StatusBadRequest)
        return
    }

    reaction := &models.Reaction{
        UserID:       userID,
        TargetType:   targetType,
        TargetID:     req.TargetID,
        ReactionType: req.ReactionType,
    }
    if err := repository.UpsertReaction(reaction); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]string{"status": "recorded"})
}

func decodeJSONBody(r *http.Request, dest interface{}) error {
    if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
        return err
    }
    return nil
}

func validateMethod(r *http.Request, allowed ...string) error {
    for _, method := range allowed {
        if r.Method == method {
            return nil
        }
    }
    return errors.New("method not allowed")
}
