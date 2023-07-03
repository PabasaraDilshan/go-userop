package main

import (
	"fmt"
	"log"
	"net/http"
	config "userop-builder/config"
	"userop-builder/userop"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func hello(w http.ResponseWriter, req *http.Request) {
    ownerAddress := common.HexToAddress( "0xAcC0FD2E9e3514DCDBcd3Ebf6144c08a44962c03")
    client, err := ethclient.Dial(config.ProviderUrl)
    if err != nil {
        log.Fatal(err)
    }
    uob :=userop.NewUserOpBuilder(ownerAddress,client)
    address := uob.GetCounterFactualAddress()
    fmt.Fprintf(w, address.String())
}

func headers(w http.ResponseWriter, req *http.Request) {

    for name, headers := range req.Header {
        for _, h := range headers {
            fmt.Fprintf(w, "%v: %v\n", name, h)
        }
    }
}

func main() {

    http.HandleFunc("/hello", hello)
    http.HandleFunc("/headers", headers)

    http.ListenAndServe(":8090", nil)
}
