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

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return &InPSQL {
		DB: db,
	}, nil
}

func (s *InPSQL) Get(key string) (string, error) {
	return "", nil
}

func (s *InPSQL) GetUserLinks(id string) (map[string]string, error){
	return nil, nil
}

func (s *InPSQL) Put(id string, key, value string) error {
	return nil
}

func (s *InPSQL) Close() error {
	s.DB.Close()
	return nil
}

func (s *InPSQL) PingDB() error {
	return s.DB.Ping()
}
