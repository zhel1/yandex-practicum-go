package storage

import (
	"database/sql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
)

type InPSQL struct {
	DB  *sql.DB
}

func NewInPSQL(databaseDSN string) (Storage, error){
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	inPSQL := InPSQL {
		DB:  db,
	}

	err = inPSQL.DB.Ping()
	if err != nil {
		log.Fatal(err)
	}

	if err = inPSQL.createTable(); err != nil {
		log.Fatal(err)
	}

	return &InPSQL {
		DB: db,
	}, nil
}

func (s *InPSQL) Get(linkID string) (string, error) {
	originURL := new(string)
	err := s.DB.QueryRow("SELECT origin_url FROM urls WHERE url_id = $1", linkID).Scan(&originURL)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return "", ErrNotFound
		default:
			log.Fatal(err)
			return "", err
		}
	}
	return "", err
}

func (s *InPSQL) GetUserLinks(userID string) (map[string]string, error){
	rows, err := s.DB.Query("SELECT url_id, origin_url FROM urls WHERE user_id = $1", userID)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	var linkID string
	var originURL string
	for rows.Next() {
		if err = rows.Scan(&linkID, &originURL); err != nil {
			return nil, err
		}
		result[linkID] = originURL
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return result, nil
}

func (s *InPSQL) Put(userID string, linkID, originURL string) error {
	_, err := s.DB.Query("INSERT INTO urls (user_id, url_id, origin_url) VALUES ($1, $2, $3)", userID, linkID, originURL)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.UniqueViolation {
			return ErrAlreadyExists
		}
		return ErrExecutionPSQL
	}
	return nil
}

func (s *InPSQL) Close() error {
	s.DB.Close()
	return nil
}

func (s *InPSQL) PingDB() error {
	return s.DB.Ping()
}
//**********************************************************************************************************************
func (s *InPSQL) createTable() error {
	query := `CREATE TABLE IF NOT EXISTS urls (
		id bigserial not null,
		user_id text not null,
		origin_url text not null,
		url_id text not null,
		UNIQUE (user_id, origin_url)
	);`
	_, err := s.DB.Exec(query)
	return err
}