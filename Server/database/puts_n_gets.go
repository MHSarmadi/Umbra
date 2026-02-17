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

func (s *BadgerStore) PutSession(ctx context.Context, u *models.Session) error {
	if len(u.UUID) == 0 {
		u.UUID = [24]byte(make([]byte, 24))
		if _, err := rand.Read(u.UUID[:]); err != nil {
			return err
		}
	}
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.ExpiresAt.IsZero() {
		u.ExpiresAt = now.Add(5 * time.Minute).UTC()
	}
	if u.LastActivity == 0 {
		u.LastActivity = now.Unix()
	}
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	err = s.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(u.KeyByUUID(), val); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (s *BadgerStore) GetSessionByUUID(ctx context.Context, uuid [24]byte) (*models.Session, error) {
	loaded := models.Session{
		UUID: uuid,
	}
	now := time.Now().UTC()
	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(loaded.KeyByUUID())
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &loaded); err != nil {
				return err
			}

			if !loaded.ExpiresAt.IsZero() && now.After(loaded.ExpiresAt) {
				if err := txn.Delete(loaded.KeyByUUID()); err != nil {
					return err
				}
				return badger.ErrKeyNotFound
			}

			// Sliding session expiration: any successful fetch extends the TTL.
			loaded.LastActivity = now.Unix()
			loaded.ExpiresAt = now.Add(5 * time.Minute).UTC()
			updated, err := json.Marshal(&loaded)
			if err != nil {
				return err
			}
			return txn.Set(loaded.KeyByUUID(), updated)
		})
	})
	if err == badger.ErrKeyNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &loaded, nil
}

func (s *BadgerStore) PutSessionInitTracker(ctx context.Context, t *models.SessionInitTracker) error {
	val, err := json.Marshal(t)
	if err != nil {
		return err
	}
	err = s.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(t.Key(), val); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (s *BadgerStore) GetSessionInitTracker(ctx context.Context, identityHash string) (*models.SessionInitTracker, error) {
	t := models.SessionInitTracker{
		IdentityHash: identityHash,
	}
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(t.Key())
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &t)
		})
	})
	if err == badger.ErrKeyNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *BadgerStore) RegisterSessionInitRequest(ctx context.Context, identityHash string, now time.Time, window time.Duration, maxRequests int, trackerTTL time.Duration) (requestCount int, limited bool, retryAfter time.Duration, err error) {
	if maxRequests <= 0 {
		return 0, false, 0, errors.New("maxRequests must be > 0")
	}
	if window <= 0 {
		return 0, false, 0, errors.New("window must be > 0")
	}
	if trackerTTL <= 0 {
		return 0, false, 0, errors.New("trackerTTL must be > 0")
	}

	tracker := models.SessionInitTracker{
		IdentityHash: identityHash,
	}

	err = s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(tracker.Key())
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
		} else {
			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &tracker)
			}); err != nil {
				return err
			}
		}

		windowStart := now.Add(-window).Unix()
		pruned := make([]int64, 0, len(tracker.RequestUnixTS)+1)
		for _, ts := range tracker.RequestUnixTS {
			if ts >= windowStart {
				pruned = append(pruned, ts)
			}
		}

		tracker.ExpiresAt = now.Add(trackerTTL).UTC()
		if len(pruned) >= maxRequests {
			limited = true
			tracker.RequestUnixTS = pruned
			if len(pruned) > 0 {
				waitSeconds := pruned[0] + int64(window.Seconds()) - now.Unix()
				if waitSeconds < 1 {
					waitSeconds = 1
				}
				retryAfter = time.Duration(waitSeconds) * time.Second
			} else {
				retryAfter = 1 * time.Second
			}
		} else {
			limited = false
			pruned = append(pruned, now.Unix())
			tracker.RequestUnixTS = pruned
			requestCount = len(pruned)
		}

		encoded, err := json.Marshal(&tracker)
		if err != nil {
			return err
		}
		return txn.Set(tracker.Key(), encoded)
	})
	if err != nil {
		return 0, false, 0, err
	}
	return requestCount, limited, retryAfter, nil
}
