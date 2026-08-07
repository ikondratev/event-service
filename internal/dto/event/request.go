package eventdto

type CreateRequest struct {
	Kind   string `json:"kind"`
	Data   string `json:"data"`
	Status string `json:"status"`
}