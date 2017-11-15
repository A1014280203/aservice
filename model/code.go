package model

import (
	"gos/dbc"
)

type Codes struct {
	
}

func (c Codes) Set(uid, code string, expSec int) {
	dbc.SetKeyValue(uid, code, expSec)
}

func (c Codes) Get(uid string) string {
	code, _ := dbc.GetKeyValue(uid)
	return code
}

func (c Codes) More(uid string, moreSec int) {
	dbc.SetKeyExpire(uid, moreSec)
}
