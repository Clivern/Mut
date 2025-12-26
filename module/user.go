// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrFailedCreateUser = errors.New("failed create user")
	ErrFailedUpdateUser = errors.New("failed update user")
	ErrFailedDeleteUser = errors.New("failed delete user")
	ErrFailedListUsers  = errors.New("failed list users")
	ErrCannotDeleteSelf = errors.New("cannot delete own account")
)

// CreateUserRequest is the body for creating a user.
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=60" label:"Name"`
	Email    string `json:"email" validate:"required,email,min=4,max=60" label:"Email"`
	Password string `json:"password" validate:"required,strong_password,min=8,max=60" label:"Password"`
	Role     string `json:"role" validate:"required,oneof=admin regular readonly" label:"Role"`
	IsActive bool   `json:"isActive" label:"Is Active"`
}

// UpdateUserRequest is the body for updating a user.
type UpdateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=60" label:"Name"`
	Email    string `json:"email" validate:"required,email,min=4,max=60" label:"Email"`
	Password string `json:"password" validate:"omitempty,strong_password,min=8,max=60" label:"Password"`
	Role     string `json:"role" validate:"required,oneof=admin regular readonly" label:"Role"`
	IsActive bool   `json:"isActive" label:"Is Active"`
}

// ListUsersResult is a page of users with the total count.
type ListUsersResult struct {
	Users []*db.User
	Total int64
}

// User handles user management.
type User struct {
	UserRepository    db.UserRepository
	SessionRepository db.SessionRepository
}

// NewUser creates a User module with the given repositories.
func NewUser(users db.UserRepository, sessions db.SessionRepository) *User {
	return &User{
		UserRepository:    users,
		SessionRepository: sessions,
	}
}

// CreateUser creates a new user.
func (u *User) CreateUser(req *CreateUserRequest) (*db.User, error) {
	existing, err := u.UserRepository.GetByEmail(req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to get user by email")
		return nil, fmt.Errorf("%w: get user: %v", ErrFailedCreateUser, err)
	}
	if existing != nil {
		return nil, ErrUserEmailAlreadyExists
	}

	password, err := util.HashPassword(req.Password)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to hash password")
		return nil, fmt.Errorf("%w: hash password: %v", ErrFailedCreateUser, err)
	}

	user := &db.User{
		Name:        req.Name,
		Email:       req.Email,
		Password:    password,
		Role:        req.Role,
		IsActive:    req.IsActive,
		LastLoginAt: time.Time{},
		Language:    db.UserLanguageEN,
		Theme:       db.UserThemeDefault,
	}

	err = u.UserRepository.Create(user)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to create user")
		return nil, fmt.Errorf("%w: create user: %v", ErrFailedCreateUser, err)
	}

	created, err := u.UserRepository.GetById(user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to reload user")
		return nil, fmt.Errorf("%w: get user: %v", ErrFailedCreateUser, err)
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("User created")

	return created, nil
}

// GetUser retrieves a user by id.
func (u *User) GetUser(id db.Id) (*db.User, error) {
	user, err := u.UserRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", id.String()).
			Msg("Failed to get user")
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateUser updates an existing user and revokes their sessions when deactivated.
func (u *User) UpdateUser(id db.Id, req *UpdateUserRequest) (*db.User, error) {
	user, err := u.GetUser(id)
	if err != nil {
		return nil, err
	}

	if req.Email != user.Email {
		existing, err := u.UserRepository.GetByEmail(req.Email)
		if err != nil {
			log.Error().
				Err(err).
				Str("email", req.Email).
				Msg("Failed to get user by email")
			return nil, fmt.Errorf("%w: get user: %v", ErrFailedUpdateUser, err)
		}
		if existing != nil && existing.Id != id {
			return nil, ErrUserEmailAlreadyExists
		}
	}

	if lo.IsNotEmpty(req.Password) {
		password, err := util.HashPassword(req.Password)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", id.String()).
				Msg("Failed to hash password")
			return nil, fmt.Errorf("%w: hash password: %v", ErrFailedUpdateUser, err)
		}

		user.Password = password
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Role = req.Role
	user.IsActive = req.IsActive

	err = u.UserRepository.Update(user)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", id.String()).
			Msg("Failed to update user")
		return nil, fmt.Errorf("%w: update user: %v", ErrFailedUpdateUser, err)
	}

	if !user.IsActive {
		err = u.SessionRepository.DeleteByUserId(id)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", id.String()).
				Msg("Failed to revoke sessions of inactive user")
		}
	}

	log.Info().
		Str("userId", id.String()).
		Msg("User updated")

	return user, nil
}

// ListUsers returns a page of users.
func (u *User) ListUsers(limit, offset int) (*ListUsersResult, error) {
	users, err := u.UserRepository.List(limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to list users")
		return nil, fmt.Errorf("%w: %v", ErrFailedListUsers, err)
	}

	total, err := u.UserRepository.Count()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to count users")
		return nil, fmt.Errorf("%w: %v", ErrFailedListUsers, err)
	}

	return &ListUsersResult{
		Users: users,
		Total: total,
	}, nil
}

// DeleteUser deletes a user other than the current one.
func (u *User) DeleteUser(id db.Id, current *db.User) error {
	if id == current.Id {
		return ErrCannotDeleteSelf
	}

	_, err := u.GetUser(id)
	if err != nil {
		return err
	}

	err = u.UserRepository.Delete(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", id.String()).
			Msg("Failed to delete user")
		return fmt.Errorf("%w: %v", ErrFailedDeleteUser, err)
	}

	log.Info().
		Str("userId", id.String()).
		Msg("User deleted")

	return nil
}
