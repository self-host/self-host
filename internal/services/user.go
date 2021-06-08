/*
Copyright (C) 2021 The Self-host Authors.
This file is part of Self-host <https://github.com/self-host/self-host>.

Self-host is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Self-host is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Self-host.  If not, see <http://www.gnu.org/licenses/>.
*/

package services

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/selfserv/rest"
	ie "github.com/self-host/self-host/internal/errors"
	pg "github.com/self-host/self-host/postgres"
)

const (
	SECRET_TOKEN_LENGTH = 40
)

// User represents the repository used for interacting with User records.
type UserService struct {
	q  *pg.Queries
	db *sql.DB
}

// NewUser instantiates the User repository.
func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		q:  pg.New(db),
		db: db,
	}
}

func (u *UserService) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	found, err := u.q.ExistsUser(ctx, id)
	if err != nil {
		return false, err
	}

	return found > 0, nil
}

func (u *UserService) AddUser(ctx context.Context, name string) (*rest.User, error) {
	// Use a transaction for this action
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		// Log?
		return nil, err
	}

	q := u.q.WithTx(tx)

	user, err := q.CreateUser(ctx, name)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else {
		tx.Commit()
	}

	ugs, err := u.q.FindGroupsByUser(ctx, user.Uuid)
	if err != nil {
		return nil, err
	}

	user_groups := make([]rest.Group, 0)
	for _, item := range ugs {
		user_groups = append(user_groups, rest.Group{
			Uuid: item.Uuid.String(),
			Name: item.Name,
		})
	}

	return &rest.User{
		Uuid:   user.Uuid.String(),
		Name:   user.Name,
		Groups: user_groups,
	}, nil
}

func (u *UserService) AddTokenToUser(ctx context.Context, user_uuid uuid.UUID, label string) (*rest.TokenWithSecret, error) {
	exists, err := u.Exists(ctx, user_uuid)
	if exists == false {
		return nil, ie.ErrorNotFound
	}

	secret := "secret-token." + RandomString(SECRET_TOKEN_LENGTH)

	params := pg.AddTokenToUserParams{
		UserUuid: user_uuid,
		Name:     label,
		Secret:   []byte(secret),
	}

	token, err := u.q.AddTokenToUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return &rest.TokenWithSecret{
		Uuid:   token.Uuid.String(),
		Name:   token.Name,
		Secret: secret,
	}, nil
}

func (u *UserService) AddRemoveUserToGroups(ctx context.Context, user_uuid uuid.UUID, adds []uuid.UUID, removes []uuid.UUID) (int64, error) {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	q := u.q.WithTx(tx)

	params := pg.RemoveUserFromGroupsParams{
		UserUuid:   user_uuid,
		GroupUuids: removes,
	}

	count, err := q.RemoveUserFromGroups(ctx, params)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for _, group_uuid := range adds {
		params := pg.AddUserToGroupParams{
			UserUuid:  user_uuid,
			GroupUuid: group_uuid,
		}

		err = q.AddUserToGroup(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	tx.Commit()

	return count, nil
}

func (u *UserService) AddUserToGroups(ctx context.Context, user_uuid uuid.UUID, group_uuids []uuid.UUID) error {
	// Use a transaction for this action
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	q := u.q.WithTx(tx)

	for _, group_uuid := range group_uuids {
		params := pg.AddUserToGroupParams{
			UserUuid:  user_uuid,
			GroupUuid: group_uuid,
		}

		err = q.AddUserToGroup(ctx, params)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()

	return nil
}

func (u *UserService) FindUserByUuid(ctx context.Context, user_uuid uuid.UUID) (*rest.User, error) {
	user, err := u.q.FindUserByUUID(ctx, user_uuid)
	if err != nil {
		return nil, err
	}

	ugs, err := u.q.FindGroupsByUser(ctx, user.Uuid)
	if err != nil {
		return nil, err
	}

	user_groups := make([]rest.Group, 0)
	for _, item := range ugs {
		user_groups = append(user_groups, rest.Group{
			Uuid: item.Uuid.String(),
			Name: item.Name,
		})
	}

	return &rest.User{
		Uuid:   user.Uuid.String(),
		Name:   user.Name,
		Groups: user_groups,
	}, nil
}

func (u *UserService) GetUserUuidFromToken(ctx context.Context, token []byte) (uuid.UUID, error) {
	id, err := u.q.GetUserUuidFromToken(ctx, token)
	if err != nil {
		return uuid.New(), err
	}
	return id, nil
}

func (u *UserService) FindAll(ctx context.Context, token []byte, limit *int64, offset *int64) ([]*rest.User, error) {
	users := make([]*rest.User, 0)

	params := pg.FindUsersParams{
		Token:     token,
		ArgLimit:  20,
		ArgOffset: 0,
	}
	if limit != nil {
		params.ArgLimit = *limit
	}
	if offset != nil {
		params.ArgOffset = *offset
	}

	user_list, err := u.q.FindUsers(ctx, params)
	if err != nil {
		return nil, err
	} else {
		for _, u := range user_list {
			var groups []rest.Group
			if err = json.Unmarshal([]byte(u.Groups), &groups); err != nil {
				// LOG error
			}

			users = append(users, &rest.User{
				Uuid:   u.Uuid.String(),
				Name:   u.Name,
				Groups: groups,
			})
		}
	}

	return users, nil
}

func (u *UserService) FindTokensForUser(ctx context.Context, user_uuid uuid.UUID) ([]*rest.Token, error) {
	tokens := make([]*rest.Token, 0)

	count, err := u.q.ExistsUser(ctx, user_uuid)
	if err != nil {
		return nil, err
	} else if count >= 0 {
		return nil, ie.ErrorNotFound
	}

	token_list, err := u.q.FindTokensByUser(ctx, user_uuid)
	if err != nil {
		return nil, err
	}

	for _, v := range token_list {
		tokens = append(tokens, &rest.Token{
			Uuid: v.Uuid.String(),
			Name: v.Name,
		})
	}

	return tokens, nil
}

func (u *UserService) RemoveUserFromGroups(ctx context.Context, user_uuid uuid.UUID, group_uuids []uuid.UUID) (int64, error) {
	params := pg.RemoveUserFromGroupsParams{
		UserUuid:   user_uuid,
		GroupUuids: group_uuids,
	}

	count, err := u.q.RemoveUserFromGroups(ctx, params)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (u *UserService) SetUserGroups(ctx context.Context, user_uuid uuid.UUID, group_uuids []uuid.UUID) (int64, error) {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	q := u.q.WithTx(tx)

	count, err := q.RemoveUserFromAllGroups(ctx, user_uuid)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for _, group_uuid := range group_uuids {
		params := pg.AddUserToGroupParams{
			UserUuid:  user_uuid,
			GroupUuid: group_uuid,
		}

		err = q.AddUserToGroup(ctx, params)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	tx.Commit()

	return count, nil
}

func (u *UserService) DeleteUser(ctx context.Context, user_uuid uuid.UUID) (int64, error) {
	count, err := u.q.DeleteUser(ctx, user_uuid)
	if err != nil {
		return 0, err
	} else if count == 0 {
		return 0, ie.ErrorNotFound
	}

	return count, nil
}

func (u *UserService) DeleteTokenFromUser(ctx context.Context, user_uuid, token_uuid uuid.UUID) (int64, error) {
	count, err := u.q.DeleteTokenFromUser(ctx, pg.DeleteTokenFromUserParams{
		UserUuid:  user_uuid,
		TokenUuid: token_uuid,
	})
	if err != nil {
		return 0, err
	} else if count == 0 {
		return 0, ie.ErrorNotFound
	}

	return count, nil
}
