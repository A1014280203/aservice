package model

import (
	"encoding/json"
)

type Exception struct {
	Stmt string `json:"error"`
}

func (e Exception) ToJSON() []byte {
	b, _ := json.Marshal(e)
	return b
}
