package cache

// Stats allows to returns some statistics of codec usage
type Stats struct {
	Hits              int `json:"hits"`
	Miss              int `json:"miss"`
	SetSuccess        int `json:"set_success"`
	SetError          int `json:"set_error"`
	DeleteSuccess     int `json:"delete_success"`
	DeleteError       int `json:"delete_error"`
	InvalidateSuccess int `json:"invalidate_success"`
	InvalidateError   int `json:"invalidate_error"`
	ClearSuccess      int `json:"clear_success"`
	ClearError        int `json:"clear_error"`
}
