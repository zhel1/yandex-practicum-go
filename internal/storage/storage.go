package storage

type DB struct {
	ShortURL map[string]string
}

func NewDBConn() *DB {
	return &DB{
		ShortURL: make(map[string]string),
	}
}
