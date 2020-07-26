package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	HashedDTO "../dto"
	"sync"
	"time"
)

//keeping all these variables in memory so that I don't have to use a "non-standard library for a DB
var hashedValuesMap sync.Map
var currentMax int = 0

func getHashedValue(w http.ResponseWriter, r *http.Request)  {
	numberString := strings.Replace(r.URL.String(), "/hash/", "", -1)
	number, err := strconv.Atoi(numberString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	returnJSON := "There is no recordId with value " + numberString

	returnObject, statusOk := hashedValuesMap.Load(number)
	if (statusOk) {
		storedObject, ok := returnObject.(HashedDTO.HashedPasswordObject)
		if (!ok) {
			fmt.Println("Unable to cast object when retrieving from Map")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		returnJSON = storedObject.HashedPassword
		if returnJSON == "" {
			returnJSON = "Value not set yet, please wait 5 seconds"
		}
	}

	fmt.Fprintf(w, returnJSON)
}

func setHashedValue(w http.ResponseWriter, r *http.Request)  {
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

func handleRequests() {
	fmt.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/hash/", getHashedValue)
	http.HandleFunc("/hash", setHashedValue)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequests()
}
