package main

import (
	"encoding/json"
	"flag"
	"log"
	"medichain/config"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"github.com/fasthttp/router"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const configPath = "config/config.json"

type Peer struct {
	PeerAddress string `json:"PeerAddress"`
}

func (p Peer) Addr() string {
	return p.PeerAddress
}

type PeerProfile struct { // connections of one peer
	ThisPeer  Peer   `json:"ThisPeer"`  // any node
	PeerPort  int    `json:"PeerPort"`  // port of peer
	Neighbors []Peer `json:"Neighbors"` // edges to that node
	Status    bool   `json:"Status"`    // Status: Alive or Dead
	Connected bool   `json:"Connected"` // If a node is connected or not [To be used later]
}

var (
	MaxPeerPort int
	PeerGraph   = make(map[string]PeerProfile)
	graphMutex  sync.RWMutex
	verbose     *bool
)

func init() {
	log.SetFlags(log.Lshortfile)
	verbose = flag.Bool("v", false, "enable verbose")
	flag.Parse()
	MaxPeerPort = 4999 // starting peer port
}

func main() {
	cfg, err := config.InitConfig(configPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	r := newRouter()
	log.Println("listening on port ", cfg.DiscoveryPort)
	go func() {
		if err := fasthttp.ListenAndServe(cfg.DiscoveryPort, r.Handler); err != nil && err != http.ErrServerClosed {
			log.Fatal("call", "ListenAndServe")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

}

func newRouter() *router.Router {
	r := router.New()
	api := r.Group("/api/v1")
	{
		api.GET("/p2p_graph", getP2pGraph)
		api.GET("/request_port", requestPeer)
		api.POST("/enroll_p2p", enrollP2p)
	}

	return r
}

func requestPeer(ctx *fasthttp.RequestCtx) {
	log.Println("handleQuery() API called")
	MaxPeerPort = MaxPeerPort + 1

	newResponse(ctx, http.StatusOK, MaxPeerPort)
	if *verbose {
		log.Println("MaxPeerPort = ", MaxPeerPort)
		spew.Dump(MaxPeerPort)
	}
}

func enrollP2p(ctx *fasthttp.RequestCtx) {
	log.Println("handleEnroll() API called")
	var incomingPeer PeerProfile

	req := ctx.Request.Body()

	if err := json.Unmarshal(req, &incomingPeer); err != nil {
		log.Println(err)
		newResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	_ = updatePeerGraph(incomingPeer)
	log.Println("Enroll request from:", incomingPeer.ThisPeer, "successful")
	newResponse(ctx, http.StatusCreated, incomingPeer)
}

func getP2pGraph(ctx *fasthttp.RequestCtx) {
	log.Println("handleQuery() API called")
	graphMutex.RLock()
	defer graphMutex.RUnlock()
	bytes, err := json.Marshal(PeerGraph)
	if err != nil {
		log.Println(err)
		newResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	newResponse(ctx, http.StatusOK, string(bytes))
	if *verbose {
		log.Println("PeerGraph = ", PeerGraph)
		spew.Dump(PeerGraph)
	}

}

func newResponse(ctx *fasthttp.RequestCtx, code int, obj interface{}) {
	strContentType := []byte("Content-Type")
	strApplicationJSON := []byte("application/json")
	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
	ctx.Response.SetStatusCode(code)

	if err := json.NewEncoder(ctx).Encode(obj); err != nil {
		zerolog.ErrorHandler(err)
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

func updatePeerGraph(inPeer PeerProfile) error {
	if *verbose {
		log.Println("incomingPeer = ", inPeer)
		spew.Dump(PeerGraph)
	}

	// Update PeerGraph
	graphMutex.Lock()
	if *verbose {
		log.Println("PeerGraph before update = ", PeerGraph)
	}
	PeerGraph[inPeer.ThisPeer.Addr()] = inPeer
	for _, neighbor := range inPeer.Neighbors {
		profile := PeerGraph[neighbor.Addr()]
		profile.Neighbors = append(profile.Neighbors, inPeer.ThisPeer)
		PeerGraph[neighbor.Addr()] = profile
	}
	if *verbose {
		log.Println("PeerGraph after update = ", PeerGraph)
		spew.Dump(PeerGraph)
	}
	graphMutex.Unlock()
	return nil
}
