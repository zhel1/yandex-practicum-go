// Package inpsql implements storage in postgres database.
package inpsql

import (
	"context"
	"errors"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	"log"
	"time"
)

// DeleteWorker is used in Storage for processing of requests for deletion
type DeleteWorker struct {
	ID  int
	st  storage.Storage
	ctx context.Context
}

// DeleteEntry is item with information for one delete operation
type DeleteEntry struct {
	UserID string
	SURL   string
}

// deleteAsyncInPSQL reads delete queue
func (d *DeleteWorker) deleteAsyncInPSQL() error {
	s, valid := d.st.(*Storage)
	if !valid {
		return errors.New("storage is not PSQL")
	}

	shutdown := false

	go func() {
		<-s.shutdown
		shutdown = true
	}()

	// listen to the channel new values and process them until chanel is closed or ctc.Done()
	for {
		select {
		case <-d.ctx.Done():
			return nil
		case records, ok := <-s.deleteQueue:
			// if channel is closed
			if !ok {
				return nil
			}

			uniqueMap := make(map[string][]string) //[user][]urls
			for _, r := range records {
				if _, exist := uniqueMap[r.UserID]; !exist {
					uniqueMap[r.UserID] = []string{r.SURL}
				} else {
					uniqueMap[r.UserID] = append(uniqueMap[r.UserID], r.SURL)
				}
			}
			for userID, sURLs := range uniqueMap {
				err := s.DeleteBatch(context.Background(), sURLs, userID)
				if err != nil {
					log.Println("DeleteBatch ERROR: ", err)
					return err
				}
			}
		default:
			if shutdown {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
