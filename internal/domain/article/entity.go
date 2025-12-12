package article

import (
	"github.com/google/uuid"
	"time"
)

type Article struct {
	ID        int64     `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	AuthorID  uuid.UUID `db:"author_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	AuthorUsername string `db:"-"`
}
