package storage

import (
	"context"
	"database/sql"
	"runtime"
	"sync"
	//_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"
	"log"
)

type InPSQL struct {
	DB  *sql.DB
	deleteQueue chan DeleteEntry
	mu  sync.Mutex
}

func NewInPSQL(databaseDSN string) (Storage, error){
	//db, err := sql.Open("pgx", databaseDSN)
	db, err := sql.Open("postgres", databaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	recordCh := make(chan DeleteEntry)

	inPSQL := InPSQL {
		DB:  db,
		deleteQueue: recordCh,
	}

	if err = inPSQL.DB.Ping(); err != nil {
		log.Fatal(err)
	}

	if err = inPSQL.createTable(); err != nil {
		log.Fatal(err)
	}

	//TODO change background context for real one
	go func() {
		g, _ := errgroup.WithContext(context.Background())
		for i := 0; i < runtime.NumCPU(); i++ {
			w := &DeleteWorker{ID: i, st: &inPSQL, ctx: context.Background()}
			g.Go(w.deleteAsyncInPSQL)
		}

		err = g.Wait()
		if err != nil {
			log.Fatal(err)
		}
	}()

	return &inPSQL, nil
}

func (s *InPSQL) Get(shortURL string) (string, error) {
	var originURL string
	var id int
	err := s.DB.QueryRow("SELECT id,origin_url FROM urls WHERE short_url = $1", shortURL).Scan(&id, &originURL)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return "", ErrNotFound
		default:
			log.Fatal(err)
			return "", err
		}
	}

	rows, err := s.DB.Query("SELECT is_deleted FROM users_url WHERE url_id = $1", id)
	if err != nil {
		return "", err
	}

	atLeastOneNotDeleted := false
	for rows.Next() {
		var isDeleted bool
		err = rows.Scan(&isDeleted)
		if err != nil {
			return "", err
		}

		if !isDeleted {
			atLeastOneNotDeleted = true
			break
		}
	}

	if !atLeastOneNotDeleted {
		return "", ErrDeleted
	}

	return originURL, err
}

func (s *InPSQL) GetUserLinks(userID string) (map[string]string, error){
	query := "SELECT short_url, origin_url FROM users_url RIGHT JOIN urls u on users_url.url_id=u.id WHERE user_id=$1;"
	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	var shortURL string
	var originURL string
	for rows.Next() {
		if err = rows.Scan(&shortURL, &originURL); err != nil {
			return nil, err
		}
		result[shortURL] = originURL
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
	}

	return result, nil
}

func (s *InPSQL) Put(userID string, shortURL, originURL string) error {
	var id int  //serial id in urls table
	query := `INSERT INTO urls (origin_url, short_url) VALUES ($1, $2) RETURNING id `
	s.DB.QueryRow(query, originURL, shortURL).Scan(&id)
	if id != 0 {
		query = `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2);`

		if _, err := s.DB.Exec(query, userID, id); err != nil {
			return err
		}
	} else {
		querySelect := `SELECT id FROM urls WHERE origin_url = $1;`
		s.DB.QueryRow(querySelect, originURL).Scan(&id)
		query = `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2) ;`

		if _, err := s.DB.Exec(query, userID, id); err != nil {
			log.Println(err)
			return ErrAlreadyExists
		}
	}
	return nil
}

func (s *InPSQL) Close() error {
	close(s.deleteQueue)
	s.DB.Close()
	return nil
}

func (s *InPSQL) PingDB() error {
	return s.DB.Ping()
}

//FanIn pattern: requests from all users are being split and put in one queue
func (s *InPSQL) Delete(shortURLs []string, userID string) error {
	var perWorkerListURL []string
	for i := 0; i < len(shortURLs); i++ {
		perWorkerListURL = append(perWorkerListURL, shortURLs[i])
		if len(perWorkerListURL) == 5 || i == len(shortURLs) - 1 {
			perWorkerBatch := DeleteEntry{UserID: userID, SURLs: perWorkerListURL}
			s.deleteQueue <- perWorkerBatch
			perWorkerListURL = []string{}
		}
	}
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
//func (s *InPSQL) createTable() error {
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
func (s *InPSQL) createTable() error {
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

