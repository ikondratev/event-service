package eventdto

import "time"

type Response struct {
	ID        int64     `json:"id"`
	Kind      string 	`json:"kind"`
	Data      string 	`json:"data"`
	Status    string 	`json:"status"`
	CreatedAt time.Time `json:"created_at"`
}