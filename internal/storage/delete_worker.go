package storage

import (
	"context"
	"errors"
	"log"
)

// DeleteWorker is used in InPSQL for processing of requests for deletion
type DeleteWorker struct {
	ID  int
	st  Storage
	ctx context.Context
}

// DeleteEntry is item with information for one delete operation
type DeleteEntry struct {
	UserID string
	SURL   string
}

// deleteAsyncInPSQL reads delete queue
func (d *DeleteWorker) deleteAsyncInPSQL() error {
	s, valid := d.st.(*InPSQL)
	if !valid {
		return errors.New("storage is not PSQL")
	}

	// listen to the channel new values and process them until chanel is closed
	for records := range s.deleteQueue {
		uniqueMap := make(map[string][]string) //[user][]urls
		for _, r := range records {
			if _, exist := uniqueMap[r.UserID]; !exist {
				uniqueMap[r.UserID] = []string{r.SURL}
			} else {
				uniqueMap[r.UserID] = append(uniqueMap[r.UserID], r.SURL)
			}
		}
		for userID, sURLs := range uniqueMap {
			err := s.DeleteBatch(sURLs, userID)
			if err != nil {
				log.Println("DeleteBatch ERROR: ", err)
				return err
			}
		}
	}
	return nil
}
