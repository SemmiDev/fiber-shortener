package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	CreateLinkTx(ctx context.Context, arg CreateLinkParams) (Link, error)
}

type SQLStore struct {
	db *sql.DB
	*Queries
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (store *SQLStore) CreateLinkTx(ctx context.Context, arg CreateLinkParams) (Link, error) {
	link := Link{ShortUrl: arg.ShortUrl, LongUrl: arg.LongUrl}

	err := store.execTx(ctx, func(q *Queries) error {
		existingLink, _ := q.GetLinkByLongURL(ctx, arg.LongUrl)
		if existingLink.LongUrl != "" {
			link = existingLink
			return nil
		}

		err := q.CreateLink(ctx, arg)
		return err
	})

	return link, err
}
