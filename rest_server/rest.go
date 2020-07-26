package main

import (
	HashedDTO "../dto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//keeping all these variables in memory so that I don't have to use a "non-standard library for a DB
var hashedValuesMap sync.Map
var currentMax int = 0
var serverRunning bool = true

func getHashedValue(w http.ResponseWriter, r *http.Request)  {
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

func getMapValue(keyNumber int, w http.ResponseWriter) HashedDTO.HashedPasswordObject {
	storedObject := HashedDTO.HashedPasswordObject{}
	returnObject, statusOk := hashedValuesMap.Load(keyNumber)
	anyError := false
	if (statusOk) {
		storedObject, statusOk = returnObject.(HashedDTO.HashedPasswordObject)
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


func setHashedValue(w http.ResponseWriter, r *http.Request)  {
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

	newBody := HashedDTO.HashedPasswordObject{RawPassword:password, CreatedTime:time.Now()}
	newKey := addToStaticList(newBody)
	go HashedDTO.HashPassword(newKey, &hashedValuesMap) //&newBody)
	fmt.Fprintf(w, strconv.Itoa(newKey))
}

//TODO: maybe synchronize this
func addToStaticList(newBody HashedDTO.HashedPasswordObject) int {
	currentMax = currentMax + 1
	hashedValuesMap.Store(currentMax, newBody)
	return currentMax
}

func readStats(w http.ResponseWriter, r *http.Request)  {
	stats := HashedDTO.Stats{}
	stats.Total = currentMax

	totalRuntime := time.Duration(0)
	for i := 1; i <= currentMax; i++ {
		returnObject := getMapValue(i, w)
		//simpliest way I could find to check for non 0
		if (returnObject.HashedTime.Year() > 0001) {
			totalRuntime = totalRuntime + (returnObject.HashedTime.Sub(returnObject.CreatedTime))
		}
	}
	avgRuntime := totalRuntime/time.Duration(stats.Total)
	stats.Average = avgRuntime/1000
	fmt.Println(stats)
	statsJson, _ := json.Marshal(stats)
	fmt.Fprintf(w, string(statsJson))
}

func prepShutdown(w http.ResponseWriter, r *http.Request)  {
	serverRunning = false
	go reallyShutdownServer()
	fmt.Fprintf(w, "Server Starting Shutdown Process")
}

func isShutdown(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprintf(w, "Server is shutting down, no new requests are being actioned")
}

func reallyShutdownServer() {
	time.Sleep(10 * time.Second)
	os.Exit(0)
}

func handleRequests() {
	fmt.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/hash/", getHashedValue)
	http.HandleFunc("/hash", setHashedValue)
	http.HandleFunc("/stats", readStats)
	http.HandleFunc("/shutdown", prepShutdown)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequests()
}
