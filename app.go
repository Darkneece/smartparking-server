package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	//for extracting service credentials from VCAP_SERVICES
	//"github.com/cloudfoundry-community/go-cfenv"
	"encoding/json"
	"io/ioutil"
)

const (
	defaultPort     = "8080"
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain"
)

var index = template.Must(template.ParseFiles(
	"templates/_base.html",
	"templates/index.html",
))

func helloworld(w http.ResponseWriter, req *http.Request) {
	index.Execute(w, nil)
}

type terminal struct {
	TerminalID string `json:"terminalID"`
	online     bool
	Address    string `json:"address"`
	Location   struct {
		Longtitude float32 `json:"longtitude"`
		Latitude   float32 `json:"latitude"`
	} `json:"location"`
}

type terminalID struct {
	TerminalID string `json:"terminalID"`
}

var terminalList []terminal

func handleTerminal(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)

	if req.Method == "POST" {
		var body terminalID
		var terminal terminal
		body.TerminalID = pseudoUUID()

		terminal.TerminalID = body.TerminalID

		bodybyte, err := json.Marshal(body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		_, err = w.Write(bodybyte)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		terminalList = append(terminalList, terminal)
	} else if req.Method == "PUT" {
		var terminal terminal
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			w.Header().Set("Content-Type", contentTypeText)
			w.Write([]byte(err.Error()))
			return
		}

		err = json.Unmarshal(body, &terminal)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", contentTypeText)
			w.Write([]byte(err.Error()))
			return
		}

		var duplicate bool

		for idx, element := range terminalList {
			if element.TerminalID == terminal.TerminalID {
				duplicate = true
				terminalList[idx] = terminal
				break
			}
		}

		if !duplicate {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", contentTypeText)
			w.Write([]byte("Invalid terminal ID"))
			return
		}
	} else if req.Method == "GET" {
		bodybyte, err := json.Marshal(terminalList)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		_, err = w.Write(bodybyte)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}

func handleHello(w http.ResponseWriter, req *http.Request) {
	if req.Method == "PUT" {
		var terminal terminal

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		err = json.Unmarshal(body, &terminal)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			w.Header().Set("Content-Type", contentTypeText)
			w.Write([]byte(err.Error()))
			return
		}

		var duplicate bool

		for idx, element := range terminalList {
			if element.TerminalID == terminal.TerminalID {
				duplicate = true
				terminalList[idx].online = true
				break
			}
		}

		if !duplicate {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", contentTypeText)
			w.Write([]byte("Invalid terminal ID"))
			return
		}

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func routes() {
	http.HandleFunc("/", helloworld)
	http.HandleFunc("/hello", handleHello)
	http.HandleFunc("/terminal", handleTerminal)
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("static"))))
}

func main() {
	var port string
	if port = os.Getenv("PORT"); len(port) == 0 {
		port = defaultPort
	}

	routes()

	log.Printf("Starting app on port %+v\n", port)
	http.ListenAndServe(":"+port, nil)
}

func pseudoUUID() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}
