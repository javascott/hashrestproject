package dto

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"time"
)

type HashedPasswordObject struct {
	RawPassword string
	HashedPassword string
	CreatedTime time.Time
	HashedTime time.Time
}

func HashPassword(passwordObject *HashedPasswordObject) {
	encryptionFunction := sha512.New()
	encryptionFunction.Write([]byte(passwordObject.RawPassword))
	passwordObject.HashedPassword = base64.URLEncoding.EncodeToString(encryptionFunction.Sum(nil))
	passwordObject.HashedTime = time.Now();
	fmt.Println(passwordObject)
}