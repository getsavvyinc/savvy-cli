package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	pgvector "github.com/pgvector/pgvector-go"

	"github.com/getsavvyinc/savvy-cli/ingest/cheatsheet"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Repo defines the interface for the database repository
type Repo interface {
	StoreExampleEmbedding(ctx context.Context, eg *cheatsheet.Example, explanationEmbedding, commandEmbedding []float32) error
}

// DB implements the Repo interface for PostgreSQL
type DB struct {
	db *sql.DB
}

var _ Repo = &DB{}

func (d *DB) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

// NewDB creates a new PostgreSQL repository
func NewDB() (*DB, error) {
	connStr := os.Getenv("PG_CONN")
	if connStr == "" {
		return nil, fmt.Errorf("PG_CONN environment variable is required")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func (pgdb *DB) StoreExampleEmbedding(ctx context.Context, eg *cheatsheet.Example, explanationEmbedding, commandEmbedding []float32) error {
	// SQL statement to insert the embeddings along with command and explanation
	query := `
	INSERT INTO command_embeddings (command, explanation, explanation_embedding, command_embedding)
	VALUES ($1, $2, $3, $4)
	`
	// Convert float32 slices to vector format using pgvector
	explanationVector := pgvector.NewVector(explanationEmbedding)
	commandVector := pgvector.NewVector(commandEmbedding)

	// Execute the query
	_, err := pgdb.db.ExecContext(ctx, query, eg.Command, eg.Explanation, explanationVector, commandVector)
	if err != nil {
		return fmt.Errorf("error executing insert query: %w", err)
	}

	return nil
}
