package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path/filepath"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

func setUpWebSite() {
	router := mux.NewRouter().StrictSlash(true)
	// Register with the WebSocket which will then push a JSON payload with data to keep the displayed data up to date. No polling necessary.
	router.HandleFunc("/ws", startDataWebSocket).Methods("GET")

	router.HandleFunc("/status", getStatus).Methods("GET")
	router.HandleFunc("/", defaultPage).Methods("GET")

	router.HandleFunc("/toggleCoil", toggleCoil).Methods("PATCH")

	fileServer := http.FileServer(neuteredFileSystem{http.Dir(webFiles)})
	router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))

	port := fmt.Sprintf(":%s", WebPort)
	log.Fatal(http.ListenAndServe(port, router))
}

func defaultPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, webFiles+"/default.html")
}

func getStatus(w http.ResponseWriter, _ *http.Request) {
	if strJson, err := Params.getJSON(); err != nil {
		fmt.Println(err)
	} else {
		if _, err := fmt.Fprintln(w, string(strJson)); err != nil {
			log.Println(err)
		}
	}
}

func toggleCoil(_ http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	address := r.FormValue("coil")
	url := fmt.Sprintf("http://aberhome1.home:8085/toggleCoil?coil=" + address)

	req, err := http.NewRequest(http.MethodPatch, url, nil) //bytes.NewBuffer(payload))
	if err != nil {
		log.Println(err)
	}
	if _, err := client.Do(req); err != nil {
		log.Println(err)
	}
}
