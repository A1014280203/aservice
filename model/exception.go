package model

import (
	"encoding/json"
)

// 暂时用不到
type Exception struct {
	Stmt string `json:"error"`
}

func (e Exception) ToJSON() []byte {
	b, _ := json.Marshal(e)
	return b
}
