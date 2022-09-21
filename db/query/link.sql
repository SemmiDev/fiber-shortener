-- name: CreateLink :exec
INSERT INTO links (short_url, long_url) VALUES ($1, $2);

-- name: GetLinkByShortURL :one
SELECT * FROM links WHERE short_url = $1;

-- name: GetLinkByLongURL :one
SELECT * FROM links WHERE long_url = $1;