// Package inpsql implements storage in postgres database.
package inpsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
	"github.com/zhel1/yandex-practicum-go/internal/storage"
	storageErrors "github.com/zhel1/yandex-practicum-go/internal/storage/errors"
	"runtime"
	"time"

	"github.com/lib/pq"

	//_ "github.com/jackc/pgx/v4/stdlib"
	"log"

	"golang.org/x/sync/errgroup"
)

//Check interface implementation
var (
	_ storage.Storage = (*Storage)(nil)
)

//DB PSQL struct
type Storage struct {
	DB          *sql.DB
	deleteBuf   chan DeleteEntry //collect url until buff is full
	deleteQueue chan []DeleteEntry
	shutdown    chan int
	done        chan int
}

//NewStorage is DB constructor
func NewStorage(databaseDSN string) (storage.Storage, error) {
	//db, err := sql.Open("pgx", databaseDSN)
	db, err := sql.Open("postgres", databaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	inPSQL := Storage{
		DB:          db,
		deleteBuf:   make(chan DeleteEntry),
		deleteQueue: make(chan []DeleteEntry),
		shutdown:    make(chan int),
		done:        make(chan int),
	}

	if err = inPSQL.DB.Ping(); err != nil {
		log.Fatal(err)
	}

	if err = inPSQL.createTable(); err != nil {
		log.Fatal(err)
	}

	go func() {
		expireTime := 5 * time.Second
		bufferSize := 5

		t := time.NewTicker(expireTime)
		parts := make([]DeleteEntry, 0, bufferSize)

		for {
			select {
			case <-t.C:
				if len(parts) > 0 {
					log.Println("Deleted URLs due to timeout")
					inPSQL.deleteQueue <- parts
					parts = make([]DeleteEntry, 0, bufferSize)
				}
			case part, ok := <-inPSQL.deleteBuf:
				if !ok { // if chanel was closed
					return
				}
				parts = append(parts, part)
				if len(parts) >= bufferSize {
					log.Println("Deleted URLs due to exceeding capacity")
					inPSQL.deleteQueue <- parts
					parts = make([]DeleteEntry, 0, bufferSize)
				}
			}
		}
	}()

	//maybe there in no sense in several go-routines, because they will not be able to write in one DB
	go func() {
		g, _ := errgroup.WithContext(context.Background())
		for i := 0; i < runtime.NumCPU(); i++ {
			w := &DeleteWorker{ID: i, st: &inPSQL, ctx: context.Background()}
			g.Go(w.deleteAsyncInPSQL)
		}

		err = g.Wait()
		inPSQL.done <- 0
		if err != nil {
			log.Fatal(err)
		}
	}()

	return &inPSQL, nil
}

//Get gets original URL from DB
func (s *Storage) Get(ctx context.Context, shortURL string) (string, error) {
	getOriginalURLStmt, err := s.DB.PrepareContext(ctx, "SELECT id,origin_url FROM urls WHERE short_url = $1;")
	if err != nil {
		return "", &storageErrors.StatementPSQLError{Err: err}
	}
	defer getOriginalURLStmt.Close()

	var originURL string
	var id int
	if err := getOriginalURLStmt.QueryRowContext(ctx, shortURL).Scan(&id, &originURL); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return "", &storageErrors.NotFoundError{Err: dto.ErrNotFound}
		default:
			return "", &storageErrors.ExecutionPSQLError{Err: err}
		}
	}

	getDeletedRowsStmt, err := s.DB.PrepareContext(ctx, "SELECT is_deleted FROM users_url WHERE url_id = $1;")
	if err != nil {
		return "", &storageErrors.StatementPSQLError{Err: err}
	}
	defer getDeletedRowsStmt.Close()

	deletedRows, err := getDeletedRowsStmt.QueryContext(ctx, id)
	if err != nil {
		return "", &storageErrors.ExecutionPSQLError{Err: err}
	}
	defer deletedRows.Close()

	atLeastOneNotDeleted := false
	for deletedRows.Next() {
		var isDeleted bool
		if err = deletedRows.Scan(&isDeleted); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return "", &storageErrors.NotFoundError{ /*Err: domain.ErrNotFound*/ }
			default:
				return "", &storageErrors.ExecutionPSQLError{Err: err}
			}
		}

		if !isDeleted {
			atLeastOneNotDeleted = true
			break
		}
	}

	err = deletedRows.Err()
	if err != nil {
		return "", &storageErrors.ExecutionPSQLError{Err: err}
	}

	if !atLeastOneNotDeleted {
		return "", dto.ErrDeleted
	}

	return originURL, nil
}

//GetUserLinks gets all URLs by UserID from DB
func (s *Storage) GetUserLinks(ctx context.Context, userID string) (map[string]string, error) {
	getUserLinksRowsStmt, err := s.DB.PrepareContext(ctx, "SELECT short_url, origin_url FROM users_url RIGHT JOIN urls u on users_url.url_id=u.id WHERE user_id=$1;")
	if err != nil {
		return nil, &storageErrors.StatementPSQLError{Err: err}
	}
	defer getUserLinksRowsStmt.Close()

	userLinksRows, err := getUserLinksRowsStmt.QueryContext(ctx, userID)
	if err != nil {
		fmt.Println(err.Error())
		return nil, &storageErrors.ExecutionPSQLError{Err: err}
	}
	defer userLinksRows.Close()

	result := make(map[string]string)
	var shortURL string
	var originURL string
	for userLinksRows.Next() {
		if err = userLinksRows.Scan(&shortURL, &originURL); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, &storageErrors.NotFoundError{ /*Err: domain.ErrNotFound*/ }
			default:
				return nil, &storageErrors.ExecutionPSQLError{Err: err}
			}
		}
		result[shortURL] = originURL
	}

	err = userLinksRows.Err()
	if err != nil {
		return nil, &storageErrors.ExecutionPSQLError{Err: err}
	}

	return result, nil
}

//Put sets short URL in DB
func (s *Storage) Put(ctx context.Context, userID string, shortURL, originURL string) error {
	id := -1 //serial id in urls table

	//add URL statement
	addURLStmt, err := s.DB.PrepareContext(ctx, `INSERT INTO urls (origin_url, short_url) VALUES ($1, $2) RETURNING id`)
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer addURLStmt.Close()

	//add user statement
	addUserStmt, err := s.DB.PrepareContext(ctx, `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2);`)
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer addUserStmt.Close()

	//get id statement
	getIDStmt, err := s.DB.PrepareContext(ctx, `SELECT id FROM urls WHERE origin_url = $1;`)
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer getIDStmt.Close()

	//begin transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}
	defer tx.Rollback()

	//add url
	//I didn't find way to use prepared statement in transaction and get id using it.
	err = addURLStmt.QueryRow(originURL, shortURL).Scan(&id)
	if err != nil {
		errCode := err.(*pq.Error).Code
		//Error "already exists" is normal for another user
		if !pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
			return &storageErrors.ExecutionPSQLError{Err: err}
		}
	}

	txAddUserStmt := tx.StmtContext(ctx, addUserStmt)
	defer txAddUserStmt.Close()

	//if new row was added
	if id != -1 {
		if _, err := txAddUserStmt.ExecContext(ctx, userID, id); err != nil {
			errCode := err.(*pq.Error).Code
			if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
				return &storageErrors.AlreadyExistsError{Err: dto.ErrAlreadyExists}
			}
			return &storageErrors.ExecutionPSQLError{Err: err}
		}
	} else { //if new row already exists
		err = getIDStmt.QueryRow(originURL).Scan(&id)
		if err != nil {
			fmt.Println(err.Error())
			return &storageErrors.ExecutionPSQLError{Err: err}
		}

		if _, err := txAddUserStmt.ExecContext(ctx, userID, id); err != nil {
			errCode := err.(*pq.Error).Code
			if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
				return &storageErrors.AlreadyExistsError{Err: dto.ErrAlreadyExists}
			}
			return &storageErrors.ExecutionPSQLError{Err: err}
		}
	}

	err = tx.Commit()
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}

	return nil
}

func (s *Storage) PutBatch(ctx context.Context, userID string, batchForDB map[string]string) error {
	id := -1 //serial id in urls table

	//add URL statement
	addURLStmt, err := s.DB.PrepareContext(ctx, `INSERT INTO urls (origin_url, short_url) VALUES ($1, $2) RETURNING id`)
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer addURLStmt.Close()

	//add user statement
	addUserStmt, err := s.DB.PrepareContext(ctx, `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2);`)
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer addUserStmt.Close()

	//get id statement
	getIDStmt, err := s.DB.PrepareContext(ctx, `SELECT id FROM urls WHERE origin_url = $1;`)
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer getIDStmt.Close()

	//begin transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}
	defer tx.Rollback()

	for originURL, shortURL := range batchForDB {
		//add url
		//I didn't find way to use prepared statement in transaction and get id using it.
		err = addURLStmt.QueryRow(originURL, shortURL).Scan(&id)
		if err != nil {
			errCode := err.(*pq.Error).Code
			//Error "already exists" is normal for another user
			if !pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
				return &storageErrors.ExecutionPSQLError{Err: err}
			}
		}

		txAddUserStmt := tx.StmtContext(ctx, addUserStmt)
		defer txAddUserStmt.Close()

		//if new row was added
		if id != -1 {
			if _, err := txAddUserStmt.ExecContext(ctx, userID, id); err != nil {
				errCode := err.(*pq.Error).Code
				if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
					return &storageErrors.AlreadyExistsError{Err: dto.ErrAlreadyExists}
				}
				return &storageErrors.ExecutionPSQLError{Err: err}
			}
		} else { //if new row already exists
			err = getIDStmt.QueryRow(originURL).Scan(&id)
			if err != nil {
				fmt.Println(err.Error())
				return &storageErrors.ExecutionPSQLError{Err: err}
			}

			if _, err := txAddUserStmt.ExecContext(ctx, userID, id); err != nil {
				errCode := err.(*pq.Error).Code
				if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
					return &storageErrors.AlreadyExistsError{Err: dto.ErrAlreadyExists}
				}
				return &storageErrors.ExecutionPSQLError{Err: err}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}

	return nil
}

//Delete deletes short URLs in DB by user ID
//It releases FanIn pattern: requests from all users are being put in one queue
func (s *Storage) Delete(ctx context.Context, shortURLs []string, userID string) error {
	for _, url := range shortURLs {
		s.deleteBuf <- DeleteEntry{UserID: userID, SURL: url}
	}
	return nil
}

//Delete deletes batch short URLs in DB by user ID
func (s *Storage) DeleteBatch(ctx context.Context, shortURLs []string, userID string) error {
	deleteStmt, err := s.DB.PrepareContext(ctx, "UPDATE users_url SET is_deleted = true WHERE user_id = $1 AND url_id = ANY(SELECT id FROM urls WHERE short_url = ANY($2));")
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}
	defer deleteStmt.Close()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}
	defer tx.Rollback()

	txDeleteStmt := tx.StmtContext(ctx, deleteStmt)
	defer txDeleteStmt.Close()

	_, err = txDeleteStmt.ExecContext(ctx, userID, pq.Array(shortURLs))
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}

	err = tx.Commit()
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}
	return nil
}

//PingDB checks connection to DB
func (s *Storage) PingDB() error {
	return s.DB.Ping()
}

//Close stops active workers disconnects from DB
func (s *Storage) Close() error {
	close(s.shutdown)
	<-s.done
	close(s.deleteBuf)
	close(s.deleteQueue)
	s.DB.Close()
	return nil
}

//**********************************************************************************************************************
// ************************************ WAITING FOR A COMMENT FROM MENTORS *********************************************
//**********************************************************************************************************************
//	There are two variant to store "is_deleted" flag:
//			   1. To store it in "urls" table. In this situation two or more users can control one URL.
//				  Problem: User can expect, that his URL can be deleted by another user. Not good. Against task.
//
//  current -> 2. To store it in "users_url" table. In this situation user can control only his copy of URL.
//				  Problem: If user delete his URL, he can expect, that his short link will continue to work.
//**********************************************************************************************************************

// 1.
//func (s *Storage) createTable() error {
//	query := `CREATE TABLE IF NOT EXISTS urls (
//		id serial primary key,
//		origin_url text not null unique,
//		short_url text not null,
//	    is_deleted boolean not null default false
//	);
//	CREATE TABLE IF NOT EXISTS users_url (
//	  user_id text not null,
//	  url_id int not null references urls(id),
//	  CONSTRAINT unique_url UNIQUE (user_id, url_id)
//	);
//	`
//	_, err := s.DB.Exec(query)
//	return err
//}
// 2.
func (s *Storage) createTable() error {
	query := `CREATE TABLE IF NOT EXISTS urls (
		id serial primary key,
		origin_url text not null unique,
		short_url text not null
	);
	CREATE TABLE IF NOT EXISTS users_url (
	  user_id text not null,
	  url_id int not null references urls(id),
	  is_deleted boolean not null default false,
	  CONSTRAINT unique_url UNIQUE (user_id, url_id)
	);
	`
	_, err := s.DB.Exec(query)
	return err
}
