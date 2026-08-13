package eventdto

type CreateRequest struct {
	Kind   string 	 `json:"kind"`
	Data   EventData `json:"data"`
	Status string 	 `json:"status"`
}

type EventData struct {
	PaymentData  PaymentData `json:"payment_data"`
	PayDevice   PayDevice `json:"payment_device"`
}

type PaymentData struct {
	AccountID string  `json:"account_id"`
	IP        string  `json:"ip_addres"`
	Location  Location `json:"location"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
	Timestamp int64   `json:"timestamp"`
	UserEmail string  `json:"user_email"`
	UserPhone string  `json:"user_phone"`
}

type Location struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type PayDevice struct {
	Token string `json:"token"`
	Type  string `json:"type"`
}