package model

import "time"

type Article struct {
	ArticleID string    `json:"articleId" bson:"_id"`   // Unique identifier for the article
	Title     string    `json:"title" bson:"title"`     // Title of the article (should be indexed)
	Content   string    `json:"content" bson:"content"` // Latest markdown content
	Version   int       `json:"version" bson:"version"` // Current version number (increments on edit)
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
}

// Revision represents a historical snapshot of an article at a specific version.
type Revision struct {
	RevisionID string    `json:"revisionId" bson:"_id"`       // Unique identifier for this specific revision
	ArticleID  string    `json:"articleId" bson:"article_id"` // Reference to the parent Article (Foreign Key)
	Version    int       `json:"version" bson:"version"`      // Version number at this point in time
	Content    string    `json:"content" bson:"content"`      // Content snapshot at this version
	Comment    string    `json:"comment" bson:"comment"`      // Summary of changes provided by the editor
	CreatedAt  time.Time `json:"createdAt" bson:"created_at"`
}

type ArticleDto struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
