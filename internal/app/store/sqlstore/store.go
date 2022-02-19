package sqlstore

import (
	"database/sql"

	"github.com/katelinlis/UserBackend/internal/app/store"
	_ "github.com/lib/pq" //db import
)

//Store ...
type Store struct {
	db                *sql.DB
	friendsRepository *FriendsRepository
	userRepository    *UserRepository
}

//New ...
func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

//Friends ...
func (s *Store) Friends() store.FriendsRepository {
	if s.friendsRepository != nil {
		return s.friendsRepository
	}

	s.friendsRepository = &FriendsRepository{
		store: s,
	}

	return s.friendsRepository
}

//User ...
func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}
