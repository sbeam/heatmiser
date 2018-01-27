package main

import (
	"fmt"
	"io/ioutil"
	"net"
)

func dial(command string) ([]byte, error) {
	host := "localhost"
	port := "9998"

	conn, err := net.Dial("tcp", host+":"+port)

	if err != nil {
		return nil, err
	} else {
		defer conn.Close()

		fmt.Fprintf(conn, "{\"command\":\""+command+"\"}")
		return ioutil.ReadAll(conn)
	}
}

func usbstats() ([]byte, error) {
	return dial("usbstats")
}

func summary() ([]byte, error) {
	return dial("summary")
}

func coin() ([]byte, error) {
	return dial("coin")
}

func main() {
	// resp, err := usbstats()
	// resp, err := summary()
	resp, err := coin()
	if err != nil {
		fmt.Println("error." + err.Error())
	} else {
		fmt.Println(fmt.Sprintf("%s", resp))
	}
}
