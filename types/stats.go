package types

// Stats allows to returns some statistics of codec usage
type Stats struct {
	Hits              int `json:"hits"`
	Miss              int `json:"miss"`
	SetSuccess        int `json:"setSuccess"`
	SetError          int `json:"setError"`
	DeleteSuccess     int `json:"deleteSuccess"`
	DeleteError       int `json:"deleteError"`
	InvalidateSuccess int `json:"invalidateSuccess"`
	InvalidateError   int `json:"invalidateError"`
	ClearSuccess      int `json:"clearSuccess"`
	ClearError        int `json:"clearError"`
}
