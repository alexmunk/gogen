package main

import (
    "fmt"
    "net/http"
    "log"
)

func doNothing(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "done")
}

func main() {
    http.HandleFunc("/", doNothing)
    err := http.ListenAndServe(":8088", nil)
    if err != nil {
        log.Fatal("Error: ", err)
    }
}