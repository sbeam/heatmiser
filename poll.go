package main

// heatmiser --rpc host:port --endpoint endpoint [summary|usbstats|coin]

import (
	"bytes"
	"crypto/tls"
	// "crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var (
	certFile  = flag.String("cert", "", "A PEM			eoncoded certificate file.")
	keyFile   = flag.String("key", "", "A PEM encoded private key file.")
	caFile    = flag.String("CA", "/path/ca_cert.pem", "A PEM eoncoded CA's certificate file.")
	rpcServer = flag.String("rpc", "host:port", "TCP location of CGMiner RPC in host:port format")
	endpoint  = flag.String("endpoint", "https://xyz.iot.us-east-1.amazonaws.com:8443", "AWS IoT API endpoint")
)

func dial(command string) []byte {
	conn, err := net.Dial("tcp", *rpcServer)

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "{\"command\":\""+command+"\"}")

	resp, err := ioutil.ReadAll(conn)

	if err != nil {
		log.Fatal(err)
	}

	// null byte terminator makes AWS's JSON parser mad
	return bytes.Trim(resp, "\x00")
}

func client() *http.Client {
	var client *http.Client
	// Setup HTTPS client
	if *certFile != "" && *keyFile != "" {
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Fatal(err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// RootCAs:      caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}
	return client
}

func publish(body []byte, topic string) (string, error) {
	// set up a POST request
	url := fmt.Sprintf("%s/topics/mining/%s?qos=1", *endpoint, topic)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	resp, err := client().Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Dump response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(data))

	return resp.Status, nil
}

func main() {
	flag.Parse()
	args := flag.Args()

	var cmd string
	if len(args) > 0 {
		cmd = args[0]
	} else {
		cmd = "summary"
	}

	resp := dial(cmd)

	fmt.Println(fmt.Sprintf("%s", resp))

	result, err := publish(resp, cmd)

	if err != nil {
		fmt.Println("error." + err.Error())
	} else {
		fmt.Println("status: " + result)
	}
}
