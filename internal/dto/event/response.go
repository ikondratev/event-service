package eventdto

import "time"

type Response struct {
	ID        int64     `json:"id"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"created_at"`
}
