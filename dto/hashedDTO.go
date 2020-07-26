package dto

import (
	"time"
)

type HashedPasswordObject struct {
	RawPassword string
	HashedPassword string
	CreatedTime time.Time
	HashedTime time.Time
}



