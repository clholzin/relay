package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	//	"sync"
	"time"
)

var (
	connType string = os.Getenv("conntype")
	ip       string = os.Getenv("hostip")
	//	wg            sync.WaitGroup
	dialConnCount int
)

func main() {
	//	wg.Add(1)
	if connType == "listen" {
		listener, err := net.Listen("tcp", ":8222")
		if err != nil {
			fmt.Errorf("Failed to create listener %v", err)
			//wg.Done()
			return
		}
		fmt.Println("Ok, listener is good")
		//go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				//wg.Done()
				fmt.Errorf("Failed to create connection %v", err)
				break
			}

			go process(conn)
		}

		//}()
	} else {

		for {
			if dialConnCount == 0 {
				conn, err := net.Dial("tcp", ip+":8222")
				if err != nil {
					fmt.Errorf("Failed to create dial conn %v - trying again", err)
					time.Sleep(time.Second)
					continue
				}
				fmt.Println("Ok, Dial conn is good")
				dialConnCount++
				go send(conn)

			}

			time.Sleep(3 * time.Second)

		}

	}

	//wg.Wait()
}

func process(c net.Conn) {
	defer fmt.Println("closing process func")
	buffer := bufio.NewScanner(c)
	var counter int
	for buffer.Scan() {
		if err := buffer.Err(); err != nil {
			fmt.Println("failed to read conn", err)
			return
		}
		counter++
		fmt.Printf("%d. receiving %s", counter, buffer.Bytes())
	}

}

func send(c net.Conn) {
	//defer wg.Done()
	defer fmt.Println("closing sender")
	data := []byte("Hello\n")
	defer func() {
		dialConnCount = 0
	}()
	var counter int
	for {
		_, err := c.Write(data[:])
		if err != nil {
			fmt.Errorf("failed to write tcp %v", err)
		}
		counter++
		fmt.Printf(" %d. sending data", counter)
		time.Sleep(time.Second)
	}

}
