package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
)

const url = "http://ipm.hermo.me/api/rest"

var isup = false

var cache ttlcache.SimpleCache = ttlcache.NewCache()

type access struct {
	Facility    string    `json:"facility"`
	Timestamp   time.Time `json:"timestamp"`
	Temperature string    `json:"temperature"`
}

func test_connection() {
	_, err := http.Head(url)
	if err != nil {
		isup = false
	} else {
		isup = true
	}
	time.Sleep(15 * time.Second)
}

func verifyCache(url string) interface{} {
	val, err := cache.Get(url)
	if err != ttlcache.ErrNotFound {
		return val
	}
	return nil
}

func auxLoopFunc(fs *[]access, k string, v interface{}, tmp map[string]interface{}) {
	format := "2006-01-02T15:04:05+04:00"
	if k == ("type") && (v == "IN") {
		var tmpacc access
		tmpacc.Temperature = tmp["temperature"].(string)
		tmpacc.Timestamp, _ = time.Parse(format, tmp["timestamp"].(string))
		var facility = tmp["facility"].(map[string]interface{})
		tmpacc.Facility = facility["name"].(string)
		*fs = append(*fs, tmpacc)
	}
}

func getAccess(rw http.ResponseWriter, r *http.Request) {
	//log.Println(isup)
	//LOW TIME NO OPT NEEDED
	var cash = verifyCache(r.URL.String())
	if cash != nil {
		var cashval = cash.(string)
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(rw, cashval)
		return
	}

	//start := time.Now()
	if !isup {
		http.Error(rw, "server unreachable", 500)
		return
	}
	//end := time.Now()
	//log.Println(end.Sub(start))

	uuids, ok := r.URL.Query()["uuid"]
	if !ok || len(uuids[0]) < 1 {
		log.Println("Url Param 'uuid' is missing")
		return
	}
	uuid := uuids[0]

	format := "2006-01-02T15:04:05+00:00"
	dt := time.Now().Add(1*time.Hour)
	dtstring := dt.Format(format)
	dt1m := dt.AddDate(-1, 0, 0).Add(1*time.Hour)
	dtm1string := dt1m.Format(format)
	//log.Println("startdate : " + dtm1string + " endate :" +dtstring)

	values := map[string]string{"startdate": dtm1string, "enddate": dtstring}
	jsonValue, _ := json.Marshal(values)
	//fmt.Fprintf(rw,string(jsonValue))

	req, err := http.NewRequest("GET", url+"/user_access_log/"+uuid+"/daterange", bytes.NewBuffer(jsonValue))
	req.Header.Set("x-hasura-admin-secret", "myadminsecretkey")

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		http.Error(rw, err.Error(), response.StatusCode)
		log.Fatal(err)
		return
	}

	var result map[string]interface{}
	//var tesresult access

	json.NewDecoder(response.Body).Decode(&result)
	//log.Println(result)


	accessarr := result["access_log"].([]interface{})
	var fs []access
	//var fspoitner = &fs

	for _, v := range accessarr {
		//fmt.Println("key", k ,"=>" , "value" , v)
		tmp := v.(map[string]interface{})
		for k, v := range tmp {
			/*
				go func() {
					if k == ("type") && (v == "IN") {
						var tmpacc access
						tmpacc.Temperature = tmp["temperature"].(string)
						tmpacc.Timestamp, _ = time.Parse(format, tmp["timestamp"].(string))
						var facility = tmp["facility"].(map[string]interface{})
						tmpacc.Facility = facility["name"].(string)
						fs = append(fs, tmpacc)
					}
				}()
			*/
			if k == ("type") && (v == "IN") {
				var tmpacc access
				tmpacc.Temperature = tmp["temperature"].(string)
				//log.Println("json " + tmp["timestamp"].(string))
				tmpacc.Timestamp, _ = time.Parse(format, tmp["timestamp"].(string))
				//log.Println("parse " + tmpacc.Timestamp.String())

				var facility = tmp["facility"].(map[string]interface{})
				tmpacc.Facility = facility["name"].(string)
				fs = append(fs, tmpacc)
			}
		}

	}

	sort.Slice(fs, func(i, j int) bool {
		return fs[i].Timestamp.After(fs[j].Timestamp)
	})
	var max = 10
	if len(fs) < 10 {
		max = len(fs)
	}

	b, _ := json.Marshal(fs)
	b, _ = json.MarshalIndent(fs[0:max], "", "  ")

	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	rw.Header().Set("Content-Type", "application/json")
	cache.Set(r.URL.String(), string(b))
	fmt.Fprintf(rw, string(b))
}

func login(rw http.ResponseWriter, r *http.Request) {

	if !isup {
		http.Error(rw, "server unreachable", 500)
		return
	}

	var u map[string]string

	err := json.NewDecoder(r.Body).Decode(&u)

	if err != nil {
		http.Error(rw, "invalid params", 400)
		return
	}
	login := u["username"]
	pwd := u["password"]

	if login == "" || pwd == "" {
		http.Error(rw, "login and password are required", 400)
		return
	}

	req, err := http.NewRequest("POST", url+"/login?username="+login+"&password="+pwd, nil)
	req.Header.Set("x-hasura-admin-secret", "myadminsecretkey")

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		http.Error(rw, err.Error(), response.StatusCode)
		log.Fatal(err)
		return
	}

	var result map[string]interface{}
	//var tesresult access

	json.NewDecoder(response.Body).Decode(&result)

	b, _ := json.Marshal(result)
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(rw, string(b))
}

func register(rw http.ResponseWriter, r *http.Request) {

	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if !isup {
		http.Error(rw, "server unreachable", 500)
		return
	}

	var u map[string]string

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(rw, "Bad params", 500)
		return
	}

	values := map[string]string{"username": u["username"], "password": u["password"], "name": u["name"], "surname": u["surname"], "phone": u["phone"], "email": u["email"], "is_vaccinated": u["is_vaccinated"]}
	jsonValue, _ := json.Marshal(values)

	req, err := http.NewRequest("POST", url+"/user", bytes.NewBuffer(jsonValue))

	req.Header.Set("x-hasura-admin-secret", "myadminsecretkey")

	client := &http.Client{}
	response, err := client.Do(req)
	//log.Println(response.StatusCode)

	if response.StatusCode == 400 {
		log.Println("error 400")
		http.Error(rw, "User already exist", response.StatusCode)
		return
	}

	if err != nil {
		http.Error(rw, err.Error(), response.StatusCode)
		return
	}
	var result map[string]interface{}
	//var tesresult access

	json.NewDecoder(response.Body).Decode(&result)
	//d,_ := json.Marshal(result)
	//fmt.Fprintf(rw, string(d))
}
func generateqr(rw http.ResponseWriter, r *http.Request) {
	var cash = verifyCache(r.URL.String())
	if cash != nil {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		rw.Header().Set("Content-Type", "image/png")
		rw.Write(cash.([]byte))
		return
	}

	names, ok := r.URL.Query()["name"]
	if !ok || len(names[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}

	surnames, ok := r.URL.Query()["surname"]
	if !ok || len(surnames[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}

	uuids, ok := r.URL.Query()["uuid"]
	if !ok || len(uuids[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}

	name := names[0]
	surname := surnames[0]
	uuid := uuids[0]
	encodeString := name + "," + surname + "," + uuid

	var png []byte

	png, err := qrcode.Encode(encodeString, qrcode.Medium, 256)

	if err != nil {
		http.Error(rw, err.Error(), 500)
		log.Fatal(err)
		return
	}
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	rw.Header().Set("Content-Type", "image/png")
	cache.Set(r.URL.String(), png)
	rw.Write(png)

}

func indexRoute(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "<h1> Welcome to IPM-API</h1><h3> methods</h3><p> /access?uuid=XXX return the last 10 access of thad uuid (GET)</p><p> /login login a username body is equals to hasura (POST)</p><p> /register register a new username body is equals to hasura (POST)</p><p> /qr?name=xx&surname=xx&uuid=xxx return an image with the name surname and uuid info</p>")

}

func main() {

	err := cache.SetTTL(15 * time.Second)
	if err != nil {
		return
	}
	go func() {
		for {
			test_connection()
		}

	}()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", indexRoute)
	router.HandleFunc("/access", getAccess).Methods("GET", "OPTIONS")
	router.HandleFunc("/login", login).Methods("POST", "OPTIONS")
	router.HandleFunc("/register", register).Methods("POST", "OPTIONS")
	router.HandleFunc("/qr", generateqr).Methods("GET", "OPTIONS")
	log.Fatal(http.ListenAndServe(":3003", router))

}
