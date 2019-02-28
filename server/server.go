package server

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"hash"
	"net/http"
	"strings"
	"sync"
	"time"
)

//The number of seconds that the server should take to respond to a /hash requst
const hashResponseTime = 5

//HashServeMux is a modified *http.ServeMux that supports a shutdown endpoint
type HashServeMux struct {
	*http.ServeMux
}

//NewHashServeMux returns a new HashServeMux
func NewHashServeMux() HashServeMux {
	return HashServeMux{http.NewServeMux()}
}

//SetUp sets up the handlers for all of the server endpoints.
//It takes as a parameter the server object to be shut down in the case of
//a call to the shutdown endpoint, and it returns a channel that will be
//closed once all in flight requests are complete in the event of a shutdown.
func (m *HashServeMux) SetUp(s *http.Server) <-chan bool {
	inFlightComplete := make(chan bool)
	shutdownTrigger := make(chan bool)
	stats := &Statistics{}

	go shutdownServerOnTrigger(s, shutdownTrigger, inFlightComplete)

	handleStatistics := newHandleStatistics(stats)
	handleShutdown := newHandleShutdown(s, shutdownTrigger)
	handleHash := newHandleHash(stats)
	m.HandleFunc("/hash", handleHash)
	m.HandleFunc("/shutdown", handleShutdown)
	m.HandleFunc("/statistics", handleStatistics)

	return inFlightComplete
}

//newHandleHash returns a HandleFunc for the /hash endpoint
func newHandleHash(stats *Statistics) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		timer := time.NewTimer(hashResponseTime * time.Second)

		//verify correct method
		if r.Method != "POST" {
			writeMethodNotAllowed(w, "POST")
			return
		}

		//calculate the encoded hash
		hasher := sha512.New()
		err := readBodyIntoHash(r, hasher)
		if err != nil {
			writeBadRequest(w)
			return
		}
		sha := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

		//write the response
		respBody := HashResponseBody{sha}
		marshalledResp, err := json.Marshal(respBody)
		if err != nil {
			writeInternalServer(w)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(marshalledResp)

		//calculate the latency
		latency := time.Now().Sub(startTime)
		latencyMs := int(latency.Nanoseconds() / 1000)
		go stats.add(latencyMs)

		//wait for the timer to expire
		<-timer.C
	}
}

//newHandleStatistics returns a HandleFunc for the /statistics endpoint
func newHandleStatistics(stats *Statistics) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writeMethodNotAllowed(w, "GET")
			return
		}

		stats.Lock()
		defer stats.Unlock()
		marshalled, err := json.Marshal(stats)
		if err != nil {
			writeInternalServer(w)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(marshalled)
	}
}

//newHandleShutdown returns a HandleFunc for the /shutdown endpoint
func newHandleShutdown(s *http.Server, triggerCh chan bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			writeMethodNotAllowed(w, "POST")
			return
		}

		triggerCh <- true
		w.WriteHeader(http.StatusAccepted)
	}
}

//readBodyIntoHash reads a request body into a hash.Hash object
func readBodyIntoHash(r *http.Request, h hash.Hash) error {
	var reqBody HashRequestBody
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	err := d.Decode(&reqBody)
	if err != nil {
		return err
	}
	h.Write([]byte(reqBody.Password))
	return nil
	/*	  var bytes []byte
	for {
	    n, err := r.Body.Read(bytes)
	    if n > 0 {
	        h.Write(bytes)
	    }
	    if err == io.EOF {
	        break
	    } else if err != nil {
	        return err
	    }
	}*/
}

//shutdownServerOnTrigger calls Shutdown on a Server when it receives a trigger
func shutdownServerOnTrigger(s *http.Server, trigger chan bool, doneCh chan bool) {
	<-trigger
	s.Shutdown(context.Background())
	close(doneCh)
}

//HashRequestBody is the request body for a request to the /hash endpoint
type HashRequestBody struct {
	Password string `json:"password"`
}

//HashResponseBody is the response body for a request to the /hash endpoint
type HashResponseBody struct {
	Hash string `json:"hash"`
}

//Statistics stores the data for the statistics endpoint
type Statistics struct {
	Total   int `json:"total"`
	Average int `json:"average"`
	sync.Mutex
}

//add adds a latency data point to the statistics
func (s *Statistics) add(t int) {
	s.Lock()
	defer s.Unlock()
	if s.Total == 0 {
		s.Average = t
		s.Total = 1
	} else {
		s.Average = ((s.Average * s.Total) + t) / (s.Total + 1)
		s.Total++
	}
}

func writeMethodNotAllowed(w http.ResponseWriter, allow ...string) {
	w.Header().Set("Allow", strings.Join(allow, ", "))
	http.Error(w, "Error: Method not allowed", http.StatusMethodNotAllowed)
}

func writeBadRequest(w http.ResponseWriter) {
	http.Error(w, "Error: Bad request.  Please consult API documentation", http.StatusBadRequest)
}

func writeInternalServer(w http.ResponseWriter, e ...error) {
	http.Error(w, "Error: An internal server error occurred", http.StatusInternalServerError)
}
