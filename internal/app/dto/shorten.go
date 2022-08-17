package dto

type UserConst string

func (c UserConst) String() string {
	return string(c)
}

var (
	UserIDCtxName UserConst = "UserID"
)

// ModelURL struct
type ModelURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ModelOriginalURLBatch struct
type ModelOriginalURLBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//ModelShortURLBatch struct
type ModelShortURLBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//ModelOriginalURL struct
type ModelOriginalURL struct {
	OriginalURL string `json:"url"`
}

//ModelShortURL struct
type ModelShortURL struct {
	ShortURL string `json:"result"`
}
