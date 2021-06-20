// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/self-host/self-host/api/aapije/rest"
	ie "github.com/self-host/self-host/internal/errors"
	"github.com/self-host/self-host/postgres"
)

const (
	secretTokenLength = 40
)

// User represents the repository used for interacting with User records.
type UserService struct {
	q  *postgres.Queries
	db *sql.DB
}

// NewUser instantiates the User repository.
func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		q:  postgres.New(db),
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
	}

	tx.Commit()

	ugs, err := u.q.FindGroupsByUser(ctx, user.Uuid)
	if err != nil {
		return nil, err
	}

	userGroups := make([]rest.Group, 0)
	for _, item := range ugs {
		userGroups = append(userGroups, rest.Group{
			Uuid: item.Uuid.String(),
			Name: item.Name,
		})
	}

	return &rest.User{
		Uuid:   user.Uuid.String(),
		Name:   user.Name,
		Groups: userGroups,
	}, nil
}

func (u *UserService) AddTokenToUser(ctx context.Context, userUUID uuid.UUID, label string) (*rest.TokenWithSecret, error) {
	exists, err := u.Exists(ctx, userUUID)
	if exists == false {
		return nil, ie.ErrorNotFound
	}

	secret := "secret-token." + RandomString(secretTokenLength)

	params := postgres.AddTokenToUserParams{
		UserUuid: userUUID,
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

func (u *UserService) AddRemoveUserToGroups(ctx context.Context, userUUID uuid.UUID, adds []uuid.UUID, removes []uuid.UUID) (int64, error) {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	q := u.q.WithTx(tx)

	params := postgres.RemoveUserFromGroupsParams{
		UserUuid:   userUUID,
		GroupUuids: removes,
	}

	count, err := q.RemoveUserFromGroups(ctx, params)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for _, groupUUID := range adds {
		params := postgres.AddUserToGroupParams{
			UserUuid:  userUUID,
			GroupUuid: groupUUID,
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

func (u *UserService) AddUserToGroups(ctx context.Context, userUUID uuid.UUID, groupUUIDs []uuid.UUID) error {
	// Use a transaction for this action
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	q := u.q.WithTx(tx)

	for _, groupUUID := range groupUUIDs {
		params := postgres.AddUserToGroupParams{
			UserUuid:  userUUID,
			GroupUuid: groupUUID,
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

func (u *UserService) FindUserByUuid(ctx context.Context, userUUID uuid.UUID) (*rest.User, error) {
	user, err := u.q.FindUserByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	ugs, err := u.q.FindGroupsByUser(ctx, user.Uuid)
	if err != nil {
		return nil, err
	}

	userGroups := make([]rest.Group, 0)
	for _, item := range ugs {
		userGroups = append(userGroups, rest.Group{
			Uuid: item.Uuid.String(),
			Name: item.Name,
		})
	}

	return &rest.User{
		Uuid:   user.Uuid.String(),
		Name:   user.Name,
		Groups: userGroups,
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

	params := postgres.FindUsersParams{
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

	userList, err := u.q.FindUsers(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, u := range userList {
		var groups []rest.Group
		if err = json.Unmarshal([]byte(u.Groups), &groups); err != nil {
			// FIXME: log error
		}

		users = append(users, &rest.User{
			Uuid:   u.Uuid.String(),
			Name:   u.Name,
			Groups: groups,
		})
	}

	return users, nil
}

func (u *UserService) FindTokensForUser(ctx context.Context, userUUID uuid.UUID) ([]*rest.Token, error) {
	tokens := make([]*rest.Token, 0)

	count, err := u.q.ExistsUser(ctx, userUUID)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, ie.ErrorNotFound
	}

	tokenList, err := u.q.FindTokensByUser(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	for _, v := range tokenList {
		tokens = append(tokens, &rest.Token{
			Uuid:    v.Uuid.String(),
			Name:    v.Name,
			Created: v.Created,
		})
	}

	return tokens, nil
}

func (u *UserService) RemoveUserFromGroups(ctx context.Context, userUUID uuid.UUID, groupUUIDs []uuid.UUID) (int64, error) {
	params := postgres.RemoveUserFromGroupsParams{
		UserUuid:   userUUID,
		GroupUuids: groupUUIDs,
	}

	count, err := u.q.RemoveUserFromGroups(ctx, params)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (u *UserService) SetUserGroups(ctx context.Context, userUUID uuid.UUID, groupUUIDs []uuid.UUID) (int64, error) {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}

	q := u.q.WithTx(tx)

	count, err := q.RemoveUserFromAllGroups(ctx, userUUID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for _, groupUUID := range groupUUIDs {
		params := postgres.AddUserToGroupParams{
			UserUuid:  userUUID,
			GroupUuid: groupUUID,
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

func (u *UserService) SetUserName(ctx context.Context, userUUID uuid.UUID, name string) (int64, error) {
	count, err := u.q.SetUserName(ctx, postgres.SetUserNameParams{
		Uuid: userUUID,
		Name: name,
	})
	if err != nil {
		return 0, err
	} else if count == 0 {
		return 0, ie.ErrorNotFound
	}

	return count, nil
}

func (u *UserService) DeleteUser(ctx context.Context, userUUID uuid.UUID) (int64, error) {
	count, err := u.q.DeleteUser(ctx, userUUID)
	if err != nil {
		return 0, err
	} else if count == 0 {
		return 0, ie.ErrorNotFound
	}

	return count, nil
}

func (u *UserService) DeleteTokenFromUser(ctx context.Context, userUUID, tokenUUID uuid.UUID) (int64, error) {
	count, err := u.q.DeleteTokenFromUser(ctx, postgres.DeleteTokenFromUserParams{
		UserUuid:  userUUID,
		TokenUuid: tokenUUID,
	})
	if err != nil {
		return 0, err
	} else if count == 0 {
		return 0, ie.ErrorNotFound
	}

	return count, nil
}
