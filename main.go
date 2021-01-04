package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
    "strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
    err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
    s.CompleteWorks = string(dat)
    s.SuffixArray = suffixarray.New([]byte(strings.ToLower(string(dat))))
	return nil
}

func (s *Searcher) Search(query string) []string {
	idxs := s.SuffixArray.Lookup([]byte(strings.ToLower(query)), -1)

	results := []string{}
	for _, idx := range idxs {
        results = append(results, s.getPhrase(idx, query))
	}
	return results
}

func (s *Searcher) getPhrase(idx int, query string) string {
    begIdx := idx
    endIdx := idx;

    testChar := string(s.CompleteWorks[idx])
    for testChar != "\n" {
        begIdx -= 1
        testChar = string(s.CompleteWorks[begIdx])
    }

    testChar = string(s.CompleteWorks[idx])
    for testChar != "." {
        endIdx += 1
        testChar = string(s.CompleteWorks[endIdx])
    }

    return s.formatPhrase(s.CompleteWorks[begIdx:endIdx+1], query)
}

func (s *Searcher) formatPhrase(p string, query string) string {
    words := strings.Split(p, " ")
    for index, word := range words {
        if (strings.ToLower(word) == strings.ToLower(query)) {
            words[index] = strings.ToUpper(word)
        }
    }

    return strings.Join(words, " ")
}


