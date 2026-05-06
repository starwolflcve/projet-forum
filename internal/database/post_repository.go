package database

import (
    "database/sql"
    "fmt"
    "strings"
    "time"

    "projet-forum/internal/models"
)

type PostRepository struct {
    DB *sql.DB
}

func NewPostRepository(db *sql.DB) (*PostRepository, error) {
    if db == nil {
        return nil, fmt.Errorf("database connection is nil")
    }

    if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
        return nil, err
    }

    return &PostRepository{DB: db}, nil
}

func (repo *PostRepository) CreatePostsTable() error {
    query := `
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    author_name TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    image_path TEXT,
    visibility TEXT NOT NULL DEFAULT 'public',
    moderation_status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
`
    _, err := repo.DB.Exec(query)
    return err
}

func (repo *PostRepository) EnsurePostAuthorNameColumn() error {
    rows, err := repo.DB.Query(`PRAGMA table_info(posts)`)
    if err != nil {
        return err
    }
    defer rows.Close()

    var exists bool
    for rows.Next() {
        var cid int
        var name string
        var ctype string
        var notnull int
        var dfltValue sql.NullString
        var pk int
        if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
            return err
        }
        if name == "author_name" {
            exists = true
            break
        }
    }

    if exists {
        return nil
    }

    _, err = repo.DB.Exec(`ALTER TABLE posts ADD COLUMN author_name TEXT NOT NULL DEFAULT ''`)
    return err
}

func (repo *PostRepository) CreateCategoriesTable() error {
    query := `
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    description TEXT
);
`
    _, err := repo.DB.Exec(query)
    return err
}

func (repo *PostRepository) CreatePostCategoriesTable() error {
    query := `
CREATE TABLE IF NOT EXISTS post_categories (
    post_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);
`
    _, err := repo.DB.Exec(query)
    return err
}

func (repo *PostRepository) CreateCommentsTable() error {
    query := `
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);
`
    _, err := repo.DB.Exec(query)
    return err
}

func (repo *PostRepository) CreateReactionsTable() error {
    query := `
CREATE TABLE IF NOT EXISTS reactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    reaction_type TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE(target_type, target_id, user_id)
);
`
    _, err := repo.DB.Exec(query)
    return err
}

func (repo *PostRepository) EnsureSchema() error {
    for _, fn := range []func() error{
        repo.CreatePostsTable,
        repo.EnsurePostAuthorNameColumn,
        repo.CreateCategoriesTable,
        repo.CreatePostCategoriesTable,
        repo.CreateCommentsTable,
        repo.CreateReactionsTable,
    } {
        if err := fn(); err != nil {
            return err
        }
    }
    return nil
}

func (repo *PostRepository) InsertPost(post *models.Post) (int, error) {
    now := time.Now().UTC()
    post.CreatedAt = now
    post.UpdatedAt = now
    result, err := repo.DB.Exec(
        `INSERT INTO posts (user_id, title, content, image_path, visibility, moderation_status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        post.UserID, post.Title, post.Content, post.ImagePath, post.Visibility, post.ModerationStatus, post.CreatedAt, post.UpdatedAt,
    )
    if err != nil {
        return 0, err
    }
    id, err := result.LastInsertId()
    return int(id), err
}

func (repo *PostRepository) UpdatePost(post *models.Post) error {
    post.UpdatedAt = time.Now().UTC()
    _, err := repo.DB.Exec(
        `UPDATE posts SET title = ?, content = ?, image_path = ?, visibility = ?, moderation_status = ?, author_name = ?, updated_at = ? WHERE id = ?`,
        post.Title, post.Content, post.ImagePath, post.Visibility, post.ModerationStatus, post.AuthorName, post.UpdatedAt, post.ID,
    )
    return err
}

func (repo *PostRepository) DeletePost(postID int) error {
    _, err := repo.DB.Exec(`DELETE FROM posts WHERE id = ?`, postID)
    return err
}

func (repo *PostRepository) InsertComment(comment *models.Comment) (int, error) {
    now := time.Now().UTC()
    comment.CreatedAt = now
    comment.UpdatedAt = now
    result, err := repo.DB.Exec(
        `INSERT INTO comments (post_id, user_id, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
        comment.PostID, comment.UserID, comment.Content, comment.CreatedAt, comment.UpdatedAt,
    )
    if err != nil {
        return 0, err
    }
    id, err := result.LastInsertId()
    return int(id), err
}

func (repo *PostRepository) UpdateComment(comment *models.Comment) error {
    comment.UpdatedAt = time.Now().UTC()
    _, err := repo.DB.Exec(
        `UPDATE comments SET content = ?, updated_at = ? WHERE id = ?`,
        comment.Content, comment.UpdatedAt, comment.ID,
    )
    return err
}

func (repo *PostRepository) DeleteComment(commentID int) error {
    _, err := repo.DB.Exec(`DELETE FROM comments WHERE id = ?`, commentID)
    return err
}

func (repo *PostRepository) InsertCategory(category *models.Category) (int, error) {
    result, err := repo.DB.Exec(
        `INSERT INTO categories (name, slug, description) VALUES (?, ?, ?)`,
        category.Name, category.Slug, category.Description,
    )
    if err != nil {
        return 0, err
    }
    id, err := result.LastInsertId()
    return int(id), err
}

func (repo *PostRepository) AttachCategoriesToPost(postID int, categoryIDs []int) error {
    tx, err := repo.DB.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()

    if _, err = tx.Exec(`DELETE FROM post_categories WHERE post_id = ?`, postID); err != nil {
        return err
    }

    stmt, err := tx.Prepare(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, categoryID := range categoryIDs {
        if _, err = stmt.Exec(postID, categoryID); err != nil {
            return err
        }
    }

    return tx.Commit()
}

func (repo *PostRepository) buildPostsQuery(filter models.PostFilter) (string, []interface{}) {
    base := `SELECT DISTINCT posts.id, posts.user_id, posts.author_name, posts.title, posts.content, posts.image_path, posts.visibility, posts.moderation_status, posts.created_at, posts.updated_at FROM posts`
    joins := []string{}
    conditions := []string{}
    args := []interface{}{}

    if filter.CategorySlug != "" {
        joins = append(joins, `JOIN post_categories ON post_categories.post_id = posts.id`)
        joins = append(joins, `JOIN categories ON categories.id = post_categories.category_id`)
        conditions = append(conditions, `categories.slug = ?`)
        args = append(args, filter.CategorySlug)
    }

    if filter.LikedByUserID > 0 {
        joins = append(joins, `JOIN reactions ON reactions.target_type = 'post' AND reactions.target_id = posts.id`)
        conditions = append(conditions, `reactions.reaction_type = 'like'`)
        conditions = append(conditions, `reactions.user_id = ?`)
        args = append(args, filter.LikedByUserID)
    }

    if filter.AuthorID > 0 {
        conditions = append(conditions, `posts.user_id = ?`)
        args = append(args, filter.AuthorID)
    }

    if filter.AuthorName != "" {
        conditions = append(conditions, `posts.author_name LIKE ?`)
        args = append(args, "%"+strings.TrimSpace(filter.AuthorName)+"%")
    }

    if filter.SearchQuery != "" {
        conditions = append(conditions, `(posts.title LIKE ? OR posts.content LIKE ?)`)
        q := "%" + strings.TrimSpace(filter.SearchQuery) + "%"
        args = append(args, q, q)
    }

    orderBy := "ORDER BY posts.created_at DESC"
    switch filter.SortBy {
    case "title":
        orderBy = "ORDER BY posts.title COLLATE NOCASE ASC"
    case "created_at":
        orderBy = "ORDER BY posts.created_at DESC"
    case "updated_at":
        orderBy = "ORDER BY posts.updated_at DESC"
    }

    query := strings.TrimSpace(base + " " + strings.Join(joins, " "))
    if len(conditions) > 0 {
        query += " WHERE " + strings.Join(conditions, " AND ")
    }
    query += " " + orderBy

    return query, args
}

func (repo *PostRepository) ListPostsWithFilters(filter models.PostFilter) ([]models.Post, error) {
    query, args := repo.buildPostsQuery(filter)
    rows, err := repo.DB.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    posts := []models.Post{}
    for rows.Next() {
        var post models.Post
        if err := rows.Scan(&post.ID, &post.UserID, &post.AuthorName, &post.Title, &post.Content, &post.ImagePath, &post.Visibility, &post.ModerationStatus, &post.CreatedAt, &post.UpdatedAt); err != nil {
            return nil, err
        }
        posts = append(posts, post)
    }
    return posts, rows.Err()
}

func (repo *PostRepository) GetPostByID(postID int) (*models.Post, error) {
    var post models.Post
    err := repo.DB.QueryRow(
        `SELECT id, user_id, author_name, title, content, image_path, visibility, moderation_status, created_at, updated_at FROM posts WHERE id = ?`,
        postID,
    ).Scan(&post.ID, &post.UserID, &post.AuthorName, &post.Title, &post.Content, &post.ImagePath, &post.Visibility, &post.ModerationStatus, &post.CreatedAt, &post.UpdatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &post, nil
}

func (repo *PostRepository) CreatePost(post *models.Post, categoryIDs []int) (int, error) {
    tx, err := repo.DB.Begin()
    if err != nil {
        return 0, err
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()

    now := time.Now().UTC()
    post.CreatedAt = now
    post.UpdatedAt = now
    result, err := tx.Exec(
        `INSERT INTO posts (user_id, author_name, title, content, image_path, visibility, moderation_status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        post.UserID, post.AuthorName, post.Title, post.Content, post.ImagePath, post.Visibility, post.ModerationStatus, post.CreatedAt, post.UpdatedAt,
    )
    if err != nil {
        return 0, err
    }

    postID, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }

    if len(categoryIDs) > 0 {
        stmt, err := tx.Prepare(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`)
        if err != nil {
            return 0, err
        }
        defer stmt.Close()
        for _, categoryID := range categoryIDs {
            if _, err := stmt.Exec(int(postID), categoryID); err != nil {
                return 0, err
            }
        }
    }

    if err := tx.Commit(); err != nil {
        return 0, err
    }
    return int(postID), nil
}

func (repo *PostRepository) ListCategories() ([]models.Category, error) {
    rows, err := repo.DB.Query(`SELECT id, name, slug, description FROM categories ORDER BY name ASC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    categories := []models.Category{}
    for rows.Next() {
        var category models.Category
        if err := rows.Scan(&category.ID, &category.Name, &category.Slug, &category.Description); err != nil {
            return nil, err
        }
        categories = append(categories, category)
    }
    return categories, rows.Err()
}

func (repo *PostRepository) ListCommentsByPostID(postID int) ([]models.Comment, error) {
    rows, err := repo.DB.Query(`SELECT id, post_id, user_id, content, created_at, updated_at FROM comments WHERE post_id = ? ORDER BY created_at ASC`, postID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    comments := []models.Comment{}
    for rows.Next() {
        var comment models.Comment
        if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
            return nil, err
        }
        comments = append(comments, comment)
    }
    return comments, rows.Err()
}

func (repo *PostRepository) CountPostLikes(postID int) (int, error) {
    return repo.countReactions("post", postID, "like")
}

func (repo *PostRepository) CountPostDislikes(postID int) (int, error) {
    return repo.countReactions("post", postID, "dislike")
}

func (repo *PostRepository) CountCommentLikes(commentID int) (int, error) {
    return repo.countReactions("comment", commentID, "like")
}

func (repo *PostRepository) CountCommentDislikes(commentID int) (int, error) {
    return repo.countReactions("comment", commentID, "dislike")
}

func (repo *PostRepository) countReactions(targetType string, targetID int, reactionType string) (int, error) {
    var count int
    err := repo.DB.QueryRow(
        `SELECT COUNT(*) FROM reactions WHERE target_type = ? AND target_id = ? AND reaction_type = ?`,
        targetType, targetID, reactionType,
    ).Scan(&count)
    return count, err
}

func (repo *PostRepository) UpsertReaction(reaction *models.Reaction) error {
    reaction.CreatedAt = time.Now().UTC()
    _, err := repo.DB.Exec(
        `INSERT INTO reactions (user_id, target_type, target_id, reaction_type, created_at) VALUES (?, ?, ?, ?, ?)
         ON CONFLICT(target_type, target_id, user_id) DO UPDATE SET reaction_type = excluded.reaction_type, created_at = excluded.created_at`,
        reaction.UserID, reaction.TargetType, reaction.TargetID, reaction.ReactionType, reaction.CreatedAt,
    )
    return err
}

func (repo *PostRepository) SeedCategories() error {
    categories := []models.Category{
        {Name: "HackTheBox", Slug: "hackthebox", Description: "Challenges et machines virtuelles HackTheBox pour pratiquer le hacking"},
        {Name: "Root Me", Slug: "rootme", Description: "Défis de sécurité Root Me - plateforme française de CTF"},
        {Name: "TryHackMe", Slug: "tryhackme", Description: "Laboratoires interactifs et défis TryHackMe"},
        {Name: "PicoCTF", Slug: "picoctf", Description: "Capture The Flag - compétition de cybersécurité"},
        {Name: "OverTheWire", Slug: "overthewire", Description: "Jeux de sécurité OverTheWire (Bandit, Natas, Leviathan)"},
        {Name: "OWASP WebGoat", Slug: "webgoat", Description: "Tutoriels de sécurité web OWASP"},
        {Name: "HtB Linux Privilege Escalation", Slug: "htb-linux-priv-esc", Description: "Escalade de privilèges Linux sur HackTheBox"},
        {Name: "HtB Windows Privilege Escalation", Slug: "htb-windows-priv-esc", Description: "Escalade de privilèges Windows sur HackTheBox"},
        {Name: "Web Exploitation", Slug: "web-exploitation", Description: "Exploitation de vulnérabilités web (SQL injection, XSS, etc)"},
        {Name: "Reverse Engineering", Slug: "reverse-engineering", Description: "Analyse et rétro-ingénierie de binaires"},
        {Name: "Cryptographie", Slug: "cryptographie", Description: "Défis de cryptographie et déchiffrement"},
        {Name: "Forensics", Slug: "forensics", Description: "Analyse médico-légale numérique et récupération de données"},
        {Name: "Steganographie", Slug: "steganographie", Description: "Cachage d'informations dans des fichiers"},
        {Name: "Recon & OSINT", Slug: "recon-osint", Description: "Reconnaissance et renseignement d'origine source ouverte"},
        {Name: "Buffer Overflow", Slug: "buffer-overflow", Description: "Exploitation de débordements de buffer"},
        {Name: "Malware Analysis", Slug: "malware-analysis", Description: "Analyse de codes malveillants"},
    }

    for _, category := range categories {
        _, err := repo.InsertCategory(&category)
        if err != nil {
            // Si la catégorie existe déjà (UNIQUE constraint), on continue
            if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
                return err
            }
        }
    }
    return nil
}

func (repo *PostRepository) ListPosts(filter models.PostFilter) ([]models.Post, error) {
    return repo.ListPostsWithFilters(filter)
}

func (repo *PostRepository) CreateComment(comment *models.Comment) (int, error) {
    return repo.InsertComment(comment)
}
