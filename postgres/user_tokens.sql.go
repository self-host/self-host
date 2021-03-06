// Code generated by sqlc. DO NOT EDIT.
// source: user_tokens.sql

package postgres

import (
	"context"

	"github.com/google/uuid"
)

const createUserToken = `-- name: CreateUserToken :one
INSERT INTO user_tokens(user_uuid, name, token_hash)
VALUES(
  $1,
  $2,
  sha256($3::bytea)
)
RETURNING uuid, user_uuid, name, token_hash, created
`

type CreateUserTokenParams struct {
	UserUuid uuid.UUID
	Name     string
	Token    []byte
}

func (q *Queries) CreateUserToken(ctx context.Context, arg CreateUserTokenParams) (UserToken, error) {
	row := q.queryRow(ctx, q.createUserTokenStmt, createUserToken, arg.UserUuid, arg.Name, arg.Token)
	var i UserToken
	err := row.Scan(
		&i.Uuid,
		&i.UserUuid,
		&i.Name,
		&i.TokenHash,
		&i.Created,
	)
	return i, err
}
