package database

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/MHSarmadi/Umbra/Server/models"
	"github.com/dgraph-io/badger/v4"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrUsernameRequired = errors.New("username required")

func (s *BadgerStore) PutUser(ctx context.Context, u *models.User) error {
	if u.Username == "" {
		return ErrUsernameRequired
	}
	if len(u.UUID) == 0 {
		u.UUID = make([]byte, 32)
		if _, err := rand.Read(u.UUID); err != nil {
			return err
		}
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	err = s.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(u.KeyByUsername()); err == nil {
			return ErrAlreadyExists
		} else if err != badger.ErrKeyNotFound {
			return err
		}
		if err := txn.Set(u.KeyByUUID(), val); err != nil {
			return err
		} else if err := txn.Set(u.KeyByUsername(), u.UUID); err != nil {
			txn.Delete(u.KeyByUUID())
			return err
		}
		return nil
	})
	return err
}

func (s *BadgerStore) GetUserByUUID(ctx context.Context, uuid []byte) (*models.User, error) {
	u := models.User{
		UUID: uuid,
	}
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(u.KeyByUUID())
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &u)
		})
	})
	if err == badger.ErrKeyNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *BadgerStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	u := models.User{
		Username: username,
	}
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(u.KeyByUsername())
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			u.UUID = append([]byte{}, val...)
			return nil
		})
	})
	if err == badger.ErrKeyNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetUserByUUID(ctx, u.UUID)
}
