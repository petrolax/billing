package transaction

type Transaction struct {
	Clientid   int64 `json:"clientid"`
	Amount     int64 `json:"amount"`
	TransferTo int64 `json:"transferto"`
	Total      int64 `json:"total"`
}
