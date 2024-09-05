package types

type GetGateQuery struct {
	ID         string
	Identifier string
}

type Gate struct {
	ID         string `json:"id" gorm:"primary_key;type:varchar(36)"`
	Identifier string `json:"identifier"`
	Complete   bool   `json:"complete"`
	Status     int    `json:"status"`
}

const (
	GateStatusCreated    = iota
	GateStatusInProgress = 1
	GateStatusSucceeded  = 2
	GateStatusFailed     = 3
)
