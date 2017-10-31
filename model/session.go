package model

import (
	// "gos/dbc"
	"time"
)

type Session struct {
	sid string
	nickname string
	createdAt  time.Time
	lastAccess time.Time

}
