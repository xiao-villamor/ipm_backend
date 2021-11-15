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
	Facility    string `json:"facility"`
	Timestamp   string `json:"timestamp"`
	Temperature string `json:"temperature"`
}
type user struct {
	username string
	password  string
}

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

	req, err := http.NewRequest("GET",url + "/user_access_log/" + uuid + "/daterange?limit=20", bytes.NewBuffer(jsonValue))
	req.Header.Set("x-hasura-admin-secret","myadminsecretkey")

	client := &http.Client{}
	response, err :=client.Do(req)

	if err != nil {
		log.Fatal(err)
		return
	}

	var result map[string]interface{}
	//var tesresult access

	json.NewDecoder(response.Body).Decode(&result)

	accessarr := result["access_log"].([]interface{})
	var fs []access

	for _,v := range accessarr {
		//fmt.Println("key", k ,"=>" , "value" , v)
		tmp := v.(map[string]interface {})
		for k,v := range tmp {
			if k == ("type") && (v == "IN") {
				var tmpacc access
				tmpacc.Timestamp = tmp["timestamp"].(string)
				tmpacc.Temperature = tmp["temperature"].(string)
				var facility = tmp["facility"].(map[string]interface {})
				tmpacc.Facility = facility["name"].(string)
				fs = append(fs, tmpacc)
			}
		}
	}

	b,_ := json.Marshal(fs)
	//fmt.Println(string(b))
	b, _ = json.MarshalIndent(fs, "", "  ")
	//log.Println(string(b))

	//fmt.Println("key", string(b))
	//log.Print(string(b))
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw,string(b))
}

func login(rw http.ResponseWriter, r *http.Request){

	var u map[string]string

	err := json.NewDecoder(r.Body).Decode(&u)

	log.Println(u)



	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	login := u["username"]
	log.Println(login)
	pwd := u["password"]
	log.Println(pwd)


	req, err := http.NewRequest("POST",url + "/login?username=" + login + "&password="+ pwd, nil)
	req.Header.Set("x-hasura-admin-secret","myadminsecretkey")

	client := &http.Client{}
	response, err :=client.Do(req)

	if err != nil {
		log.Fatal(err)
		return
	}

	var result map[string]interface{}
	//var tesresult access

	json.NewDecoder(response.Body).Decode(&result)

	b,_ := json.Marshal(result)
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw,string(b))
}

func register(rw http.ResponseWriter, r *http.Request){

	var u map[string]string

	err := json.NewDecoder(r.Body).Decode(&u)

	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	log.Println(u)

	values  := map[string]string{"username": u["username"], "password": u["password"], "name": u["name"], "surname": u["surname"], "phone": u["phone"], "email": u["email"], "is_vaccinated": u["is_vaccinated"]}
	jsonValue,_ := json.Marshal(values)


	req, err := http.NewRequest("POST",url+"/user", bytes.NewBuffer(jsonValue))
	req.Header.Set("x-hasura-admin-secret","myadminsecretkey")

	client := &http.Client{}
	response, err :=client.Do(req)

	if err != nil {
		log.Fatal(err)
		return
	}

	var result map[string]interface{}
	//var tesresult access

	json.NewDecoder(response.Body).Decode(&result)

	//d,_ := json.Marshal(result)
	//fmt.Fprintf(rw, string(d))
}

func indexRoute(rw http.ResponseWriter, r *http.Request){
	fmt.Fprintf(rw,"welcome to contacts API")

}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/",indexRoute)
	router.HandleFunc("/access/{id}",getAccess).Methods("GET")
	router.HandleFunc("/login",login).Methods("POST")
	router.HandleFunc("/register",register).Methods("POST")
	log.Fatal(http.ListenAndServe(":3003",router))

}
