package database

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/MHSarmadi/Umbra/Server/models"
	"github.com/dgraph-io/badger/v4"
)

func (s *BadgerStore) StartExpiryJanitor(ctx context.Context, sweepInterval time.Duration) {
	if sweepInterval <= 0 {
		sweepInterval = 1 * time.Minute
	}

	// Sweep immediately on startup so restarts clean stale data.
	if removedSessions, removedTrackers, err := s.SweepExpired(ctx, time.Now().UTC()); err != nil {
		log.Printf("expiry janitor initial sweep error: %v", err)
	} else {
		log.Printf("expiry janitor initial sweep removed sessions=%d trackers=%d", removedSessions, removedTrackers)
	}

	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			removedSessions, removedTrackers, err := s.SweepExpired(ctx, t.UTC())
			if err != nil {
				log.Printf("expiry janitor sweep error: %v", err)
				continue
			}
			if removedSessions > 0 || removedTrackers > 0 {
				log.Printf("expiry janitor removed sessions=%d trackers=%d", removedSessions, removedTrackers)
			}
		}
	}
}

func (s *BadgerStore) SweepExpired(ctx context.Context, now time.Time) (removedSessions int, removedTrackers int, err error) {
	err = s.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		// Session keys are prefix 0x10 and exactly 25 bytes (1 prefix + 24 UUID bytes).
		for it.Seek([]byte{0x10}); it.ValidForPrefix([]byte{0x10}); it.Next() {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			item := it.Item()
			key := item.KeyCopy(nil)
			if len(key) != 25 {
				continue
			}

			var session models.Session
			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &session)
			}); err != nil {
				continue
			}

			if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
				if err := txn.Delete(key); err != nil {
					return err
				}
				removedSessions++
			}
		}

		// Session-init tracker keys are prefix 0x12.
		for it.Seek([]byte{0x12}); it.ValidForPrefix([]byte{0x12}); it.Next() {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			item := it.Item()
			key := item.KeyCopy(nil)

			var tracker models.SessionInitTracker
			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &tracker)
			}); err != nil {
				continue
			}

			if !tracker.ExpiresAt.IsZero() && now.After(tracker.ExpiresAt) {
				if err := txn.Delete(key); err != nil {
					return err
				}
				removedTrackers++
			}
		}

		return nil
	})
	return removedSessions, removedTrackers, err
}
