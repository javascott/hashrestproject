package controllerClasses

import (
	DTOs "../dto"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//keeping all these variables in memory so that I don't have to use a "non-standard library for a DB
var hashedValuesMap sync.Map
var currentMax int = 0
var serverRunning bool = true
const hashThreadSleepTime = 15
//This seems to always be 3 after all threads finish... I couldn't find more in the runtime library
var InitialGoThreads = 3

//GET Function to read value from static Map
func GetHashedValue(w http.ResponseWriter, r *http.Request)  {
	numberString := strings.Replace(r.URL.String(), "/hash/", "", -1)
	number, err := strconv.Atoi(numberString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	returnJSON := "There is no recordId with value " + numberString
	if (number <= currentMax) {
		returnObject := getMapValue(number, w)
		returnJSON = returnObject.HashedPassword
		if returnJSON == "" {
			returnJSON = "Value not set yet, please wait 5 seconds"
		}
	}

	fmt.Fprintf(w, returnJSON)
}

//POST function to store into static Map
func SetHashedValue(w http.ResponseWriter, r *http.Request)  {
	if (!serverRunning) {
		isShutdown(w, r)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	const passwordPrefix = "password="

	//checking POST Body
	//TODO: ask if I need to check case sensativity
	validBody := true
	if (strings.Index(string(body), passwordPrefix) < 0) {
		//bad request body format
		validBody = false
	}
	password := strings.Replace(string(body), passwordPrefix, "", -1)
	if (password == "") {
		validBody = false
	}
	if (!validBody) {
		http.Error(w, "Invalid request body.  Syntax should be password=xxxxx", http.StatusBadRequest)
		return
	}

	newBody := DTOs.HashedPasswordObject{RawPassword:password, CreatedTime:time.Now()}
	newKey := addToStaticList(newBody)
	go HashPassword(newKey, &hashedValuesMap) //&newBody)
	fmt.Fprintf(w, strconv.Itoa(newKey))
}

//TODO: maybe synchronize this
func addToStaticList(newBody DTOs.HashedPasswordObject) int {
	currentMax = currentMax + 1
	hashedValuesMap.Store(currentMax, newBody)
	return currentMax
}


//function for thread to hash password after 5 seconds
func HashPassword(newKey int, hashedValuesMap *sync.Map) {
	time.Sleep(hashThreadSleepTime * time.Second)
	mapObject, statusOk := hashedValuesMap.Load(newKey)
	if (statusOk) {
		passwordObject, ok := mapObject.(DTOs.HashedPasswordObject)
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

//Util function to get value from Map since you have to load and cast
func getMapValue(keyNumber int, w http.ResponseWriter) DTOs.HashedPasswordObject {
	storedObject := DTOs.HashedPasswordObject{}
	returnObject, statusOk := hashedValuesMap.Load(keyNumber)
	anyError := false
	if (statusOk) {
		storedObject, statusOk = returnObject.(DTOs.HashedPasswordObject)
		if (!statusOk) {
			anyError = true
		}
	} else {
		anyError = true
	}
	if (anyError) {
		http.Error(w, "Unable to cast object when retrieving from Map", http.StatusInternalServerError)
	}
	return storedObject
}


//status functions
func ReadStats(w http.ResponseWriter, r *http.Request)  {
	stats := DTOs.Stats{}
	stats.Total = currentMax

	totalRuntime := time.Duration(0)
	for i := 1; i <= currentMax; i++ {
		returnObject := getMapValue(i, w)
		//simpliest way I could find to check for non 0 time
		if (returnObject.HashedTime.Year() > 0001) {
			totalRuntime = totalRuntime + (returnObject.HashedTime.Sub(returnObject.CreatedTime))
		}
	}
	avgRuntime := totalRuntime/time.Duration(stats.Total)
	stats.Average = avgRuntime/1000
	statsJson, _ := json.Marshal(stats)
	fmt.Fprintf(w, string(statsJson))
}


//shutdown Controller functions
func PrepShutdown(w http.ResponseWriter, r *http.Request)  {
	serverRunning = false
	go reallyShutdownServer()
	fmt.Fprintf(w, "Server Starting Shutdown Process")
}

func isShutdown(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprintf(w, "Server is shutting down, no new requests are being actioned")
}

func reallyShutdownServer() {
	for runtime.NumGoroutine() > InitialGoThreads {
		time.Sleep(1 * time.Second)
	}
	os.Exit(0)
}


