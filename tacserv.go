package main

import (
	"code.google.com/p/gosqlite/sqlite"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func log(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	log(time.Now().UTC().Format("2006-01-02 15:04:05.000"), "["+r.RemoteAddr+"]", r.Method, r.URL.String())
	bail := func(status int, message string) {
		w.WriteHeader(status)
		w.Write([]byte(message))
	}
	serverError := func(message string) {
		bail(http.StatusInternalServerError, message)
	}
	respond := func(body interface{}, err error) {
		if err != nil {
			serverError(err.Error())
		} else {
			w.Write([]byte(fmt.Sprint(body)))
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	urlparts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	nodeId64, err := strconv.ParseInt(urlparts[0], 10, 64)
	nodeId := int(nodeId64)
	if err != nil {
		bail(http.StatusNotFound, "Invalid node")
		return
	}
	if len(urlparts) < 2 {
		bail(http.StatusNotFound, "Resource not found")
		return
	}

	operation := urlparts[1]
	switch operation {
	case "card":
		switch r.Method {
		case "GET":
			respond(getPermissions(nodeId, urlparts[2]))
		case "POST":
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}
			body := strings.TrimSpace(string(bodyBytes))
			cards := strings.Split(body, ",")
			if len(cards) != 2 {
				serverError("malformed request -- expected \"card_id,granter_card_id\"")
				return
			}
			respond(setPermissions(nodeId, cards[0], cards[1], 1))
		}
	case "sync":
		if len(urlparts) > 3 {
			bail(http.StatusNotFound, "no such resource")
			return
		}
		lastCardId := ""
		if len(urlparts) == 3 {
			lastCardId = urlparts[2]
		}
		cardId, err := getNextCard(nodeId, lastCardId)
		if err != nil {
			serverError(err.Error())
			return
		}
		if cardId == "" {
			bail(http.StatusNotFound, "no more card ids")
			return
		}
		respond(cardId, nil)
	default:
		bail(http.StatusNotFound, "no such resource")
	}
}

var db *sqlite.Conn

func main() {
	dbfile := new(string)
	flag.StringVar(dbfile, "db", "tac.sqlite", "Filename of sqlite3 database for persistent data.")
	port := new(int)
	flag.IntVar(port, "port", 8080, "Port for http server to listen on.")
	flag.Parse()

	log("Using database file", *dbfile)
	var err error
	db, err = sqlite.Open(*dbfile)
	if err != nil {
		log(err)
		os.Exit(1)
	}
	err = setup(db)
	if err != nil {
		log(err)
		os.Exit(1)
	}
	http.HandleFunc("/", requestHandler)
	log("Starting server on port", *port)
	log(http.ListenAndServe(":"+strconv.FormatInt(int64(*port), 10), nil))
}
