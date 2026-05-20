package store

import (
	"log"
	"time"
)

func StartGC(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := runGC(); err != nil {
				log.Printf("[GC] error: %v", err)
			}
		}
	}()
}

func runGC() error {
	now := time.Now()

	res, err := DB.Exec(
		`DELETE FROM refresh_tokens WHERE revoked = 1 OR expires_at < ?`, now,
	)
	if err != nil {
		return err
	}
	refreshDeleted, _ := res.RowsAffected()

	res, err = DB.Exec(
		`DELETE FROM authorization_codes WHERE used = 1 OR expires_at < ?`, now,
	)
	codesDeleted, _ := res.RowsAffected()

	if refreshDeleted > 0 || codesDeleted > 0 {
		log.Printf("[GC] cleaned up %d refresh token(s), %d authorization code(s)",
			refreshDeleted, codesDeleted,
		)
	}

	return nil
}
