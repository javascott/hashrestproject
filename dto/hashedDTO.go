package dto

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

type HashedPasswordObject struct {
	RawPassword string
	HashedPassword string
	CreatedTime time.Time
	HashedTime time.Time
}

type Stats struct {
	Total int
	Average time.Duration
}

func HashPassword(newKey int, hashedValuesMap *sync.Map) { //passwordObject *HashedPasswordObject) {
	time.Sleep(5 * time.Second)
	mapObject, statusOk := hashedValuesMap.Load(newKey)
	if (statusOk) {
		passwordObject, ok := mapObject.(HashedPasswordObject)
		if (!ok) {
			fmt.Println("Unable to cast object when retrieving from Map")
		}
		encryptionFunction := sha512.New()
		encryptionFunction.Write([]byte(passwordObject.RawPassword))
		passwordObject.HashedPassword = base64.URLEncoding.EncodeToString(encryptionFunction.Sum(nil))
		passwordObject.HashedTime = time.Now();
		//TODO: figure out how to do an "update" in sync.Map, documentation wasn't the best... but delete and restore works.
		hashedValuesMap.Delete(newKey)
		hashedValuesMap.Store(newKey, passwordObject)
	}

}