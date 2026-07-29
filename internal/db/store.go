package db

import (
	"context"
	"database/sql"

	"github.com/danny-pearson/word-association/internal/db/gen"
)

type Store struct {
	db      *sql.DB
	queries *gen.Queries
}

func NewStore(conn *sql.DB) *Store {
	return &Store{
		db:      conn,
		queries: gen.New(conn),
	}
}

func (s *Store) RecentAnswers(ctx context.Context, limit int64) ([]string, error) {
	return s.queries.RecentAnswers(ctx, limit)
}

func (s *Store) FindExistingAnswers(ctx context.Context, answers []string) ([]string, error) {
	return s.queries.FindExistingAnswers(ctx, answers)
}
