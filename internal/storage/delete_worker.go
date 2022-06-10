package storage

import (
	"context"
	"errors"
	"github.com/lib/pq"
	"log"
)

type DeleteWorker struct {
	ID  int
	st  Storage
	ctx context.Context
}

type DeleteEntry struct {
	UserID string
	SURLs  []string
}

func (d *DeleteWorker) deleteAsyncInPSQL() error {
	s, valid := d.st.(*InPSQL)
	if !valid {
		return errors.New("storage is not PSQL")
	}

	// prepare DELETE statement
	deleteStmt, err := s.DB.PrepareContext(d.ctx, "UPDATE users_url SET is_deleted = true WHERE user_id = $1 AND url_id = ALL(SELECT id FROM urls WHERE short_url = ANY($2))")
	if err != nil {
		return err //err
	}
	defer deleteStmt.Close()

	// begin transaction
	tx, err := s.DB.BeginTx(d.ctx, nil)
	if err != nil {
		return ErrExecutionPSQL //err
	}
	defer tx.Rollback()
	txDeleteStmt := tx.StmtContext(d.ctx, deleteStmt)

	// listen to the channel new values and process them until chanel is closed
	for record := range s.deleteQueue {
		s.mu.Lock()
		_, err = txDeleteStmt.ExecContext(d.ctx, record.UserID, pq.Array(record.SURLs))
		if err != nil {
			s.mu.Unlock()
			return ErrExecutionPSQL //err
		}
		log.Println("Worker ID ", d.ID, "Deleting URL ", record.SURLs)
		err := tx.Commit()
		if err != nil {
			s.mu.Unlock()
			return ErrExecutionPSQL
		}
		s.mu.Unlock()
	}
	return nil
}