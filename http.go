package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strings"
)

type users struct {
	Id    uint64 `json:"id,omitempty"`
	Name  string `json:"name"`
	Utype string `json:"type,omitempty"`
}

type relationships struct {
	No       string `json:"no,omitempty"`
	Mainuser string `json:"user_id,omitempty"`
	Tuser    string `json:" user_id,omitempty"`
	Liked    string `json:"liked,omitempty"`
	Matched  string `json:"matched,omitempty"`
	State    string `json:"state,omitempty"`
	Rtype    string `json:"type,omitempty"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "tantan"
)

func OpenConnection() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

var (
	hostname string
	curport  int
)

/* register command line options */
func init() {
	flag.StringVar(&hostname, "localhost", "0.0.0.0", "The hostname or IP on which the REST server will listen")
	flag.IntVar(&curport, "port", 80, "The port on which the REST server will listen")
}
func main() {
	flag.Parse()
	var address = fmt.Sprintf("%s:%d", hostname, curport)
	log.Println("REST service listening on", address)

	router := mux.NewRouter().StrictSlash(true)
	router.
		HandleFunc("/users", GetHandler).
		Methods("GET")

	router.
		HandleFunc("/users", PostHandler).
		Methods("POST")

	router.
		HandleFunc("/users/{Id}/relationships", rGETHandler).
		Methods("GET")
	router.
		HandleFunc("/users/{userid1}/relationships/{userid2}", rPUTHandler).
		Methods("PUT")

	err := http.ListenAndServe(address, router)
	if err != nil {
		log.Fatalln("ListenAndServe err:", err)
	}

	log.Println("Server end")
	//log.Fatal(http.ListenAndServe(":80", nil))
}

func rPUTHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	re := relationships{Rtype: "relationship"}
	ok := true
	err := json.NewDecoder(r.Body).Decode(&re)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		ok = false
	}
	vars := mux.Vars(r)
	userid1 := vars["userid1"]
	userid2 := vars["userid2"]
	re.Mainuser = userid1
	re.Tuser = userid2
	var cur_is_l string
	var cur_is_m string
	if strings.EqualFold(re.State, "matched") {
		cur_is_m = "1"
		cur_is_l = "1"
	}
	if strings.EqualFold(re.State, "liked") {
		cur_is_l = "1"
		cur_is_m = "0"
	}
	if strings.EqualFold(re.State, "disliked") {
		cur_is_m = "0"
		cur_is_l = "0"
	}
	rows, err := db.Query("SELECT * FROM relationship r WHERE (r.mainuser= $1 and r.touser=$2)", userid1, userid2)
	if err != nil {
		log.Fatal(err)
		ok = false
	}
	var s string
	var t string
	err2 := db.QueryRow("SELECT mainuser,touser FROM relationship r WHERE r.mainuser= $1 and r.touser=$2", userid1, userid2).Scan(&s, &t)
	if err2 == sql.ErrNoRows {
		//"There is not row"
		if strings.EqualFold(cur_is_m, "1") {
			sqlStatement1 := `INSERT INTO relationship (mainuser,touser,liked,is_matched) VALUES ($1,$2,$3,$4)`
			_, err = db.Exec(sqlStatement1, re.Mainuser, re.Tuser, cur_is_l, cur_is_m)
			if err != nil {
				panic(err)
				ok = false
			}
			//w.WriteHeader(http.StatusOK)
			sqlStatement2 := `INSERT INTO relationship (mainuser,touser,liked,is_matched) VALUES ($1,$2,$3,$4)`
			_, err = db.Exec(sqlStatement2, re.Tuser, re.Mainuser, cur_is_l, cur_is_m)
			if err != nil {
				ok = false
				panic(err)
			}
		} else {

			sqlStatement3 := `INSERT INTO relationship (mainuser,touser,liked,is_matched) VALUES ($1,$2,$3,$4)`
			_, err = db.Exec(sqlStatement3, re.Mainuser, re.Tuser, cur_is_l, cur_is_m)
			if err != nil {
				panic(err)
				ok = false
			}

		}
	} else {
		for rows.Next() { //there are rows
			fmt.Println("there are rows")
			var is_l string
			var is_m string
			err := rows.Scan(&re.Mainuser, &re.Tuser, &re.Liked, &re.Matched, &re.No)
			if err != nil {
				fmt.Println(err.Error())
				ok = false
			}
			is_l = re.Liked
			is_m = re.Matched
			if strings.EqualFold(is_m, "1") { //两用户本来就match
				fmt.Println(re.State)
				if strings.EqualFold(re.State, "disliked") { //user1单方面dislike user2
					cur_is_m = "0"
					cur_is_l = "0"
					stmt, err := db.Prepare("UPDATE relationship set liked=$1,is_matched=$2 where mainuser=$3 and touser=$4")
					if err != nil {
						log.Fatal(err)
						ok = false
					}
					_, err = stmt.Exec(cur_is_l, cur_is_m, re.Mainuser, re.Tuser)
					if err != nil {
						log.Fatal(err)
						ok = false
					}

					stmt2, err := db.Prepare("UPDATE relationship set liked=$1,is_matched=$2 where mainuser=$3 and touser=$4")
					if err != nil {
						log.Fatal(err)
						ok = false
					}
					_, err = stmt2.Exec(cur_is_l, cur_is_m, re.Mainuser, re.Tuser)
					if err != nil {
						log.Fatal(err)
						ok = false
					}

				}

			} else {
				if strings.EqualFold(is_l, "1") {
					if strings.EqualFold(re.State, "dislike") {
						cur_is_m = "0"
						cur_is_l = "0"
						stmt3, err := db.Prepare("UPDATE relationship set liked=$1,is_matched= $2 where mainuser=$3")
						if err != nil {
							log.Fatal(err)
							ok = false
						}
						_, err = stmt3.Exec(cur_is_l, cur_is_m, re.Mainuser)
						if err != nil {
							log.Fatal(err)
							ok = false
						}
						stmt4, err := db.Prepare("UPDATE relationship set is_matched= $1 where mainuser=$2")
						if err != nil {
							ok = false
							log.Fatal(err)

						}
						_, err = stmt4.Exec(cur_is_m, re.Tuser)
						if err != nil {
							ok = false
						}

					}
					if strings.EqualFold(re.State, "matched") {
						cur_is_m = "1"
						cur_is_l = "1"
						stmt3, err := db.Prepare("UPDATE relationship set liked=$1,is_matched= $2 where mainuser=$3")
						if err != nil {
							log.Fatal(err)
							ok = false
						}
						_, err = stmt3.Exec(cur_is_l, cur_is_m, re.Mainuser)
						if err != nil {

							ok = false
						}

						stmt4, err := db.Prepare("UPDATE relationship set liked=$1,is_matched= $2 where mainuser=$3")
						if err != nil {
							log.Fatal(err)
							ok = false
						}
						_, err = stmt4.Exec(cur_is_l, cur_is_m, re.Tuser)
						if err != nil {

							ok = false
						}

					}
				} else {
					if strings.EqualFold(re.State, "matched") {
						cur_is_m = "1"
						cur_is_l = "1"
						stmt3, err := db.Prepare("UPDATE relationship set liked=$1,is_matched= $2 where mainuser=$3")
						if err != nil {
							log.Fatal(err)
							ok = false
						}
						_, err = stmt3.Exec(cur_is_l, cur_is_m, re.Mainuser)
						if err != nil {

							ok = false
						}

						stmt4, err := db.Prepare("UPDATE relationship set liked=$1,is_matched= $2 where mainuser=$3")
						if err != nil {
							log.Fatal(err)
							ok = false
						}
						_, err = stmt4.Exec(cur_is_l, cur_is_m, re.Tuser)
						if err != nil {
							log.Fatal(err)
							ok = false
						}

					}
					if strings.EqualFold(re.State, "liked") {
						cur_is_m = "0"
						cur_is_l = "1"
						stmt3, err := db.Prepare("UPDATE relationship set liked=$1,is_matched= $2 where mainuser=$3")
						if err != nil {

							ok = false
						}
						_, err = stmt3.Exec(cur_is_l, cur_is_m, re.Mainuser)
						if err != nil {
							log.Fatal(err)
							ok = false
						}

					}

				}
			}
		}

	}
	re.Matched = ""
	re.Liked = ""
	re.No = ""
	re.Mainuser = ""
	var res map[string]string = make(map[string]string)
	var status = http.StatusOK

	if !ok {
		res["result"] = "fail"
		res["error"] = "required parameter name is missing"
		status = http.StatusBadRequest
	} else {
		res["result"] = "success"
		res["name"] = "put new relationship in the database"
		status = http.StatusOK
	}
	st, _ := json.Marshal(status)
	response, _ := json.Marshal(res)

	usersBytes, _ := json.MarshalIndent(re, "", "\t")

	w.Header().Set("Content-Type", "application/json,charset=utf-8")

	w.Write(usersBytes)
	w.Write(st)
	w.Write(response)

	defer rows.Close()
	defer db.Close()

}

func rGETHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	vars := mux.Vars(r)
	userid := vars["Id"]
	ok := true
	rows, err := db.Query("SELECT * FROM relationship WHERE mainuser = $1 ", userid)
	if err != nil {
		log.Fatal(err)
		ok = false
	}
	var relist []relationships

	for rows.Next() {
		re := relationships{Rtype: "relationship"}
		var l string
		var m string
		var n string
		rows.Scan(&re.Mainuser, &re.Tuser, l, m, n)

		if strings.EqualFold(m, "1") {
			re.State = "matched"

		} else {
			if strings.EqualFold(l, "1") {
				re.State = "liked"
			} else {
				re.State = "disliked"
			}
		}
		re.Mainuser = ""
		relist = append(relist, re)
	}

	usersBytes, _ := json.MarshalIndent(relist, "", "\t")

	w.Header().Set("Content-Type", "application/json")

	var res map[string]string = make(map[string]string)
	var status = http.StatusOK

	if !ok {
		res["result"] = "fail"
		res["error"] = "required parameter name is missing"
		status = http.StatusBadRequest
	} else {
		res["result"] = "success"
		res["name"] = "finding all of the user`s relationship"
	}
	st, _ := json.Marshal(status)
	response, _ := json.Marshal(res)

	w.Header().Set("Content-Type", "application/json")
	w.Write(usersBytes)
	w.Write(st)
	w.Write(response)

	defer rows.Close()
	defer db.Close()
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	ok := true
	var u users
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
		ok = false
	}

	sqlStatement := `INSERT INTO users (name) VALUES ($1)`
	_, err = db.Exec(sqlStatement, u.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
		ok = false
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	rows, err := db.Query("SELECT * FROM users WHERE name=  $1 order by id desc limit 1 ", u.Name)
	if err != nil {
		log.Fatal(err)
		ok = false
	}

	for rows.Next() {
		//var name string

		adduser := users{Utype: "user"}
		rows.Scan(&adduser.Id, &adduser.Name)
		usersBytes, _ := json.MarshalIndent(adduser, "", "\t")
		w.Write(usersBytes)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
		ok = false
	}
	var res map[string]string = make(map[string]string)
	var status = http.StatusOK

	if !ok {
		res["result"] = "fail"
		res["error"] = "required parameter name is missing"
		status = http.StatusBadRequest
	} else {
		res["result"] = "succ"
		res["name"] = "put new user in the database"
	}
	st, _ := json.Marshal(status)
	response, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(st)
	w.Write(response)
	defer rows.Close()
	defer db.Close()
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	ok := true
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}

	var ulist []users

	for rows.Next() {
		u := users{Utype: "user"}
		rows.Scan(&u.Id, &u.Name)

		ulist = append(ulist, u)
	}

	usersBytes, _ := json.MarshalIndent(ulist, "", "\t")

	var res map[string]string = make(map[string]string)
	var status = http.StatusOK


	if !ok {
		res["result"] = "fail"
		res["error"] = "required parameter name is missing"
		status = http.StatusBadRequest
	} else {
		res["result"] = "success"
		res["name"] = "get all of the users"

	}
	st, _ := json.Marshal(status)
	response, _ := json.Marshal(res)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(usersBytes)
	w.Write(st)
	w.Write(response)

	defer rows.Close()
	defer db.Close()

}
