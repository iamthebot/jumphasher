package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/iamthebot/jumphasher/common"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

//Central API engine
//
//Responsible for dispatching work, etc.
type APIEngine struct {
	store    jumphasher.HashStore   //holds our hashes
	metrics  MetricsEngine          //keeps track of metrics
	inChans  []chan *HashingRequest //used to route hashing requests to our workers
	alive    jumphasher.AtomicFlag  //used to coordinate shutdown
	sslcfg   *SSLConfig             //ssl configuration. If nil, SSL is disabled
	port     int                    //port to listen on
	delay    int                    //number of seconds to delay hashing requests
	hashType int                    //hashing engine to use
	wg       sync.WaitGroup         //used to coordinate shutdown for workers
}

//Initializes a new API engine
//
//c: Desired concurrency
//
//hf: Hash function
//
//sslcfg: SSL/TLS configuration if applicable
//
//port: Port to listen on
//
//delay: Number of seconds to delay each hashing request
func NewAPIEngine(c int, hf int, sslcfg *SSLConfig, port int, delay int) (*APIEngine, error) {
	var e APIEngine
	e.inChans = make([]chan *HashingRequest, c)
	e.alive.Clear()
	e.sslcfg = sslcfg
	e.hashType = hf
	e.port = port
	e.delay = delay
	return &e, nil
}

func (e *APIEngine) Start() {
	for i := 0; i < len(e.inChans); i++ {
		e.inChans[i] = make(chan *HashingRequest)
	}

	//set up handlers for default muxer
	http.HandleFunc("/hash", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			e.onHashGet(w, req)
		case "POST":
			e.onHashPost(w, req)
		default:
			http.Error(w, fmt.Sprintf("Unsupported method: %s", req.Method), 405)
		}
	})
	http.HandleFunc("/stats", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, fmt.Sprintf("Unsupported method: %s", req.Method), 405)
			return
		}
		e.onStatsGet(w, req)
	})
	http.HandleFunc("/shutdown", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, fmt.Sprintf("Unsupported method: %s", req.Method), 405)
			return
		}
		e.onShutdownGet(w, req)
	})
	if e.sslcfg != nil {
		e.alive.TestAndSet()
		if !e.sslcfg.Exclusive { //dispatch TLS listener asynchronously
			go func() {
				log.Printf("Server now accepting SSL connections at port %d", e.sslcfg.Port)
				err := http.ListenAndServeTLS(fmt.Sprintf(":%d", e.sslcfg.Port), e.sslcfg.CertFile, e.sslcfg.KeyFile, nil)
				if err != nil {
					log.Fatal(err)
				}
			}()
		} else { //block on TLS listener
			log.Printf("Server now accepting SSL connections at port %d", e.port)
			err := http.ListenAndServeTLS(fmt.Sprintf(":%d", e.sslcfg.Port), e.sslcfg.CertFile, e.sslcfg.KeyFile, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		e.alive.TestAndSet()
		log.Printf("Server now accepting http connections at port %d", e.port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", e.port), nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}

//Gracefully shut down API engine
//
//First, declares a shutdown state so further requests are rejected
//
//Finally, waits for workers to finish
func (e *APIEngine) Stop() {
	//enable the shutdown flag to prevent conflicts
	e.alive.Clear()
	//add a short timeout to allow open HTTP responses to complete
	time.Sleep(1)
	//close worker channels
	for _, c := range e.inChans {
		close(c)
	}
	//wait for workers to finish
	e.wg.Wait()
}

//Handles incoming work requests for hashing and dispatches async persistence tasks
func (e *APIEngine) worker(c chan *HashingRequest) {
	e.wg.Add(1)
	var he jumphasher.HashingEngine
	switch e.hashType { //we can extend this with more hash functions
	case jumphasher.HashTypeSHA512:
		he = jumphasher.NewSHA512Engine()
	default:
		log.Fatal("Unknown hash function")
	}
	for r := range c {
		//hash the request
		h, err := he.Hash(r.Password)
		if err != nil {
			r.ReturnChan <- err
		}
		//dispatch async persistence job
		go e.persist(&r.ID, h, e.delay)
		r.ReturnChan <- nil
	}
	e.wg.Done()
}

//Persists hashing result asynchronously
//
//delay controls the number of seconds before the ID is persisted and the result is available via GET /hash
func (e *APIEngine) persist(id *jumphasher.UUID, hash []byte, delay int) {
	if delay != 0 {
		time.Sleep(time.Duration(int64(delay) * int64(time.Second)))
	}
	err := e.store.Store(id, hash)
	if err != nil {
		log.Printf("Error: job %s could not be stored", id.MarshalText())
	}
}

//route handler for POST /hash
func (e *APIEngine) onHashPost(w http.ResponseWriter, req *http.Request) {
	//if the API is shutting down, we need to immediately return a 503
	if !e.alive.Test() {
		http.Error(w, "server is shutting down", http.StatusServiceUnavailable)
		return
	}
	start := time.Now()
	defer req.Body.Close()
	password, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rc := make(chan error)

	//generate Job ID
	id, err := jumphasher.UUIDv4()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//figure out where to route the request
	worker_id := binary.LittleEndian.Uint32(id[0:4]) % uint32(len(e.inChans))
	r := HashingRequest{
		ID:         *id,
		Password:   password,
		ReturnChan: rc,
	}
	e.inChans[worker_id] <- &r

	//wait on the response
	err = <-rc
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	strid := id.MarshalText()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(strid)))
	io.WriteString(w, strid)
	elapsed := time.Since(start)
	e.metrics.AddDuration(elapsed.Nanoseconds())
}

//route handler for GET /hash
func (e *APIEngine) onHashGet(w http.ResponseWriter, req *http.Request) {
	strid := req.URL.Query().Get("id")
	if strid == "" {
		http.Error(w, "must provide a job ID via the 'id' parameter", http.StatusBadRequest)
		return
	}
	var u jumphasher.UUID
	err := u.UnmarshalText(strid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//look up hash
	hash, err := e.store.Load(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if hash == nil {
		r := fmt.Sprintf("hash for job id %s not found", strid)
		http.Error(w, r, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	base64hash := base64.StdEncoding.EncodeToString(hash)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(base64hash)))
}

//route handler for GET /stats
func (e *APIEngine) onStatsGet(w http.ResponseWriter, req *http.Request) {
	//fetch metrics snapshot
	snap := e.metrics.MSSnapshot()
	j, err := snap.MarshalJson()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(j)))
	w.Write(j)
}

//route handler for GET /shutdown
func (e *APIEngine) onShutdownGet(w http.ResponseWriter, req *http.Request) {
	if !e.alive.Test() {
		http.Error(w, "server already shutting down", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	r := "commencing shutdown"
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(r)))
	io.WriteString(w, r)
	log.Println("Received shutdown request. Commencing shutdown")
	go e.Stop()
}
