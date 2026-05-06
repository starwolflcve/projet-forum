package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "projet-forum/internal/models"
)

type commentRequest struct {
    PostID  int    `json:"post_id"`
    Content string `json:"content"`
}

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req commentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON payload", http.StatusBadRequest)
        return
    }

    userID, err := parseUserID(r)
    if err != nil || userID <= 0 {
        http.Error(w, "missing or invalid user_id", http.StatusBadRequest)
        return
    }

    comment := &models.Comment{
        PostID:  req.PostID,
        UserID:  userID,
        Content: req.Content,
    }

    commentID, err := repository.CreateComment(comment)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]interface{}{"id": commentID})
}

func UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    if r.Method != http.MethodPut && r.Method != http.MethodPatch {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    commentID, err := parseIDQuery(r, "id")
    if err != nil || commentID <= 0 {
        http.Error(w, "missing or invalid comment id", http.StatusBadRequest)
        return
    }

    var req commentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON payload", http.StatusBadRequest)
        return
    }

    comment := &models.Comment{
        ID:      commentID,
        Content: req.Content,
    }

    if err := repository.UpdateComment(comment); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]string{"status": "updated"})
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
    if repository == nil {
        http.Error(w, "repository not initialized", http.StatusInternalServerError)
        return
    }

    if r.Method != http.MethodDelete {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    commentID, err := parseIDQuery(r, "id")
    if err != nil || commentID <= 0 {
        http.Error(w, "missing or invalid comment id", http.StatusBadRequest)
        return
    }

    if err := repository.DeleteComment(commentID); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]string{"status": "deleted"})
}

func ReactToCommentHandler(w http.ResponseWriter, r *http.Request) {
    reactHandler(w, r, "comment")
}
