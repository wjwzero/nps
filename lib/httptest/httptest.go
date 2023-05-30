package main

import (
	"fmt"
	"io"
	"net/http"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("r.Method = ", r.Method)
	fmt.Println("r.URL = ", r.URL)
	fmt.Println("r.Header = ", r.Header)
	fmt.Println("r.Body = ", r.Body)
	fmt.Println("============================================")

	//回复
	io.WriteString(w, "0!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
}

func main() {
	http.HandleFunc("/control/pull/subscriber", IndexHandler)
	http.ListenAndServe("0.0.0.0:5212", nil)
}
