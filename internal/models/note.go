package models

import (
	"time"
)

type Note struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateNoteRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=255"`
	Content string `json:"content" binding:"required,min=1"`
}

type UpdateNoteRequest struct {
	Title   string `json:"title" binding:"omitempty,min=1,max=255"`
	Content string `json:"content" binding:"omitempty,min=1"`
}

type PaginationParams struct {
	Limit      int       `form:"limit" binding:"omitempty,min=1,max=100"`
	CursorTime time.Time `form:"cursor_time" binding:"omitempty"`
	CursorID   int64     `form:"cursor_id" binding:"omitempty,min=0"`
}

type NotesPage struct {
	Notes    []Note `json:"notes"`
	NextPage bool   `json:"next_page"`
	Cursor   string `json:"cursor,omitempty"`
	Total    int64  `json:"total,omitempty"`
}
