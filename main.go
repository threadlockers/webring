package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"strings"

	_ "embed"

	"github.com/syumai/workers"
)

type WebringEntry struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

//go:embed webring.json
var webringRaw []byte
var hostsToIgnore = []string{"ring.seggs.lol", "seggs.lol", "www.seggs.lol"}

func main() {
	var webring []WebringEntry
	if err := json.Unmarshal(webringRaw, &webring); err != nil {
		slog.Error("failed to unmarshal webring json file", "error", err)
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host

		if strings.HasSuffix(host, ".seggs.lol") && !slices.Contains(hostsToIgnore, host) {
			sub := strings.TrimSuffix(host, ".seggs.lol")
			for _, entry := range webring {
				if entry.Name == sub {
					http.Redirect(w, r, entry.Url, http.StatusFound)
					return
				}
			}

			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		setCorsHeaders(w)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("uwu"))
	})

	http.HandleFunc("/webring", func(w http.ResponseWriter, r *http.Request) {
		buildJsonResponse(w, http.StatusOK, webring)
	})

	http.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("from")
		if from == "" {
			buildJsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "missing `from` query parameter",
			})
			return
		}

		dir := r.URL.Query().Get("dir")
		if dir == "" {
			buildJsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "missing `dir` query parameter",
			})
			return
		}

		if dir != "next" && dir != "prev" {
			buildJsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "invalid `dir` query parameter. it can be either `next` or `prev` only",
			})
			return
		}

		index := -1
		for i, v := range webring {
			if v.Name == from {
				index = i
				break
			}
		}

		if index == -1 {
			buildJsonResponse(w, http.StatusBadRequest, map[string]string{
				"error": "invalid `from` query parameter. can't find any webring entry's name as `from`",
			})
			return
		}

		url := ""

		if dir == "prev" {
			if index == 0 {
				url = webring[len(webring)-1].Url
			} else {
				url = webring[index-1].Url
			}
		} else {
			if index == len(webring)-1 {
				url = webring[0].Url
			} else {
				url = webring[index+1].Url
			}
		}

		setCorsHeaders(w)
		http.Redirect(w, r, url, http.StatusFound)
	})

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		setCorsHeaders(w)
		index := rand.Intn(len(webring))
		http.Redirect(w, r, webring[index].Url, http.StatusFound)
	})

	workers.Serve(nil)
	fmt.Println("server is up and running at :8080")
}

func setCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func buildJsonResponse(w http.ResponseWriter, statusCode int, v any) {
	setCorsHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}
