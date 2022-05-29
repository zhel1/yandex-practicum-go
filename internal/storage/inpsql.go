package storage

import (
	"database/sql"
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

func (s *InPSQL) Get(shortURL string) (string, error) {
	var originURL string
	err := s.DB.QueryRow("SELECT origin_url FROM urls WHERE short_url = $1", shortURL).Scan(&originURL)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return "", ErrNotFound
		default:
			log.Fatal(err)
			return "", err
		}
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
	s.DB.Close()
	return nil
}

func (s *InPSQL) PingDB() error {
	return s.DB.Ping()
}
//**********************************************************************************************************************
func (s *InPSQL) createTable() error {
	query := `CREATE TABLE IF NOT EXISTS urls (
		id serial primary key,
		origin_url text not null unique,
		short_url text not null 
	);
	CREATE TABLE IF NOT EXISTS users_url(
	  user_id text not null ,
	  url_id int not null  references urls(id),
	  CONSTRAINT unique_url UNIQUE (user_id, url_id)
	);
	`
	_, err := s.DB.Exec(query)
	return err
}