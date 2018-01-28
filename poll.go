package main

// heatmiser --rpc host:port --endpoint endpoint [summary|usbstats|coin]

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var (
	certFile  = flag.String("cert", "", "Path to a PEM encoded certificate file.")
	keyFile   = flag.String("key", "", "Path to a PEM encoded private key file.")
	caFile    = flag.String("CA", "", "Path to a PEM encoded CA's certificate file.")
	rpcServer = flag.String("rpc", "", "location of CGMiner RPC server over TCP [host:port]")
	endpoint  = flag.String("endpoint", "", "AWS IoT API endpoint [https://host:port]")
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
	var tlsConfig *tls.Config

	// Setup HTTPS client
	if *certFile != "" && *keyFile != "" {
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Fatal(err)
		}

		// Load CA cert
		if *caFile != "" {
			caCert, err := ioutil.ReadFile(*caFile)
			if err != nil {
				log.Fatal(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      caCertPool,
			}
		} else {
			tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
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
	if *endpoint == "" {
		log.Fatal("API recieving endpoint must be set with --endpoint")
	}
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

	var cmd string
	if flag.NArg() > 0 {
		cmd = flag.Args()[0]
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
