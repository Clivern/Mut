// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"github.com/clivern/mut/db"
)

type fakeUsers struct {
	db.UserRepository
	byId      map[db.Id]*db.User
	created   []*db.User
	updated   []*db.User
	deleted   []db.Id
	lastLogin []db.Id
	err       error
}

func (f *fakeUsers) Create(user *db.User) error {
	if f.err != nil {
		return f.err
	}

	user.Id = db.Id("u-new")
	f.created = append(f.created, user)
	if f.byId == nil {
		f.byId = map[db.Id]*db.User{}
	}
	f.byId[user.Id] = user

	return nil
}

func (f *fakeUsers) GetById(id db.Id) (*db.User, error) {
	return f.byId[id], f.err
}

func (f *fakeUsers) GetByEmail(email string) (*db.User, error) {
	for _, user := range f.byId {
		if user.Email == email {
			return user, f.err
		}
	}

	return nil, f.err
}

func (f *fakeUsers) Update(user *db.User) error {
	f.updated = append(f.updated, user)

	return f.err
}

func (f *fakeUsers) UpdateLastLogin(id db.Id) error {
	f.lastLogin = append(f.lastLogin, id)

	return f.err
}

func (f *fakeUsers) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)
	delete(f.byId, id)

	return f.err
}

func (f *fakeUsers) List(_, _ int) ([]*db.User, error) {
	list := []*db.User{}
	for _, user := range f.byId {
		list = append(list, user)
	}

	return list, f.err
}

func (f *fakeUsers) Count() (int64, error) {
	return int64(len(f.byId)), f.err
}

type fakeSessions struct {
	db.SessionRepository
	created       []*db.Session
	deletedByUser []db.Id
}

func (f *fakeSessions) Create(session *db.Session) error {
	session.Id = db.Id("s-new")
	f.created = append(f.created, session)

	return nil
}

func (f *fakeSessions) GetByToken(token string) (*db.Session, error) {
	for _, session := range f.created {
		if session.Token == token {
			return session, nil
		}
	}

	return nil, nil
}

func (f *fakeSessions) DeleteByUserId(userId db.Id) error {
	f.deletedByUser = append(f.deletedByUser, userId)

	return nil
}

func (f *fakeSessions) DeleteExpired() (int64, error) {
	return 0, nil
}

type fakeConfigs struct {
	db.ConfigRepository
	values map[string]string
	err    error
}

func (f *fakeConfigs) Create(key, value string) error {
	f.values[key] = value

	return f.err
}

func (f *fakeConfigs) Get(key string) (*db.Config, error) {
	if f.err != nil {
		return nil, f.err
	}

	value, ok := f.values[key]
	if !ok {
		return nil, nil
	}

	return &db.Config{Key: key, Value: value}, nil
}

func (f *fakeConfigs) Update(key, value string) error {
	f.values[key] = value

	return f.err
}
