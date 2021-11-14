package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"time"
)

const url = "http://ipm.hermo.me/api/rest"

type access struct {
	Facility    string `json:Facility`
	Date        string `json:Date`
	Temperature string `json:Temperature`
}

type allAccess []access

func getAccess (rw http.ResponseWriter,r *http.Request) {

	vars := mux.Vars(r)
	uuid := vars["id"]
	format := "2006-01-02T15:04:05+03:00"
	dt := time.Now()
	dtstring := dt.Format(format)
	dt1m := dt.AddDate(-1,0,0)
	dtm1string := dt1m.Format(format)


	values  := map[string]string{"startdate": dtm1string, "enddate": dtstring}
	jsonValue,_ := json.Marshal(values)
	//fmt.Fprintf(rw,string(jsonValue))

	req, err := http.NewRequest("GET","http://ipm.hermo.me/api/rest/user_access_log/" + uuid + "/daterange?limit=10", bytes.NewBuffer(jsonValue))
	req.Header.Set("x-hasura-admin-secret","myadminsecretkey")

	client := &http.Client{}
	response, err :=client.Do(req)

	if err != nil {
		log.Fatal(err)
		return
	}

	var result map[string]interface{}

	json.NewDecoder(response.Body).Decode(&result)
    b,_ := json.Marshal(result)
	log.Print(result)
	fmt.Fprintf(rw,string(b))
	}


func indexRoute(rw http.ResponseWriter, r *http.Request){
	fmt.Fprintf(rw,"welcome to contacts API")

}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/",indexRoute)
	router.HandleFunc("/access/{id}",getAccess).Methods("GET")
	log.Fatal(http.ListenAndServe(":3000",router))

}