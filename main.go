package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dkrisa/jumpcloud/server"
)

func main() {
	fmt.Println("Hello, Jumpcloud.")
	defer fmt.Println("Goodbye, Jumpcloud.")

	mux := server.NewHashServeMux()
	s := &http.Server{
		Addr:           ":8000",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	inFlightComplete := mux.SetUp(s)
	log.Println("INFO:", s.ListenAndServe())
	<-inFlightComplete
}

func shutdownServerOnTrigger(s *http.Server, c chan bool, doneCh chan bool) {
	<-c
	s.Shutdown(context.Background())
	doneCh <- true
}

func newHandleStatistics(stats *Statistics) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		stats.Lock()
		defer stats.Unlock()
		marshalled, err := json.Marshal(stats)
		if err != nil {
			//do something
		}
		w.Write(marshalled)
	}
}

func newHandleHash(stats *Statistics) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		timer := time.NewTimer(5 * time.Second)

		var reqBody HashRequestBody
		d := json.NewDecoder(r.Body)
		d.DisallowUnknownFields()
		err := d.Decode(&reqBody)
		if err != nil {
			//internal server error
		}
		hasher := sha512.New()
		hasher.Write([]byte(reqBody.Password))
		sha := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
		respBody := HashResponseBody{sha}
		marshalledResp, err := json.Marshal(respBody)
		w.WriteHeader(http.StatusOK)
		w.Write(marshalledResp)
		latency := time.Now().Sub(startTime)
		latencyMs := int(latency.Nanoseconds() / 1000)
		go stats.Add(latencyMs)
		fmt.Println(latencyMs)
		<-timer.C
	}
}

func newHandleShutdown(s *http.Server, triggerCh chan bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		//write something
		triggerCh <- true
	}
}

func xhandleHash(w http.ResponseWriter, r *http.Request) {
	/*h := sha512.New()
	  var bytes []byte
	  for {
	      n, err := r.Body.Read(bytes)
	      if n > 0 {
	          h.Write(bytes)
	      }
	      if err == io.EOF {
	          break
	      } else if err != nil {
	          //write internal server error
	      }
	  }*/
	timer := time.NewTimer(5 * time.Second)

	var reqBody HashRequestBody
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	err := d.Decode(&reqBody)
	if err != nil {
		//internal server error
	}
	hasher := sha512.New()
	hasher.Write([]byte(reqBody.Password))
	fmt.Println(reqBody.Password)
	sha := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	fmt.Println(sha)
	respBody := HashResponseBody{sha}
	marshalledResp, err := json.Marshal(respBody)
	w.WriteHeader(http.StatusOK)
	w.Write(marshalledResp)
	<-timer.C
}

type HashRequestBody struct {
	Password string `json:"password"`
}

type HashResponseBody struct {
	Hash string `json:"hash"`
}

func (s *Statistics) Add(t int) {
	s.Lock()
	defer s.Unlock()
	if s.Total == 0 {
		s.Average = t
		s.Total = 1
	} else {
		s.Average = ((s.Average * s.Total) + t) / (s.Total + 1)
		s.Total += 1
	}
}

type Statistics struct {
	Total   int `json:"total"`
	Average int `json:"average"`
	sync.Mutex
}
