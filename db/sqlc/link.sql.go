// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: link.sql

package db

import (
	"context"
)

const createLink = `-- name: CreateLink :exec
INSERT INTO links (short_url, long_url) VALUES ($1, $2)
`

type CreateLinkParams struct {
	ShortUrl string `json:"short_url"`
	LongUrl  string `json:"long_url"`
}

func (q *Queries) CreateLink(ctx context.Context, arg CreateLinkParams) error {
	_, err := q.db.ExecContext(ctx, createLink, arg.ShortUrl, arg.LongUrl)
	return err
}

const getLinkByLongURL = `-- name: GetLinkByLongURL :one
SELECT short_url, long_url FROM links WHERE long_url = $1
`

func (q *Queries) GetLinkByLongURL(ctx context.Context, longUrl string) (Link, error) {
	row := q.db.QueryRowContext(ctx, getLinkByLongURL, longUrl)
	var i Link
	err := row.Scan(&i.ShortUrl, &i.LongUrl)
	return i, err
}

const getLinkByShortURL = `-- name: GetLinkByShortURL :one
SELECT short_url, long_url FROM links WHERE short_url = $1
`

func (q *Queries) GetLinkByShortURL(ctx context.Context, shortUrl string) (Link, error) {
	row := q.db.QueryRowContext(ctx, getLinkByShortURL, shortUrl)
	var i Link
	err := row.Scan(&i.ShortUrl, &i.LongUrl)
	return i, err
}
