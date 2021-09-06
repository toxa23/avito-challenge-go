package app

type Balance struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type InputPayload struct {
	Amount  int64  `json:"amount"`
	Details string `json:"details,omitempty"`
	UserId  int64  `json:"user_id,omitempty"`
}
