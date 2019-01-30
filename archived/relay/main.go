package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

var (
	ip = os.Getenv("hubip")
)

func main() {
	//	wg.Add(1)
	fmt.Println("relay")

	var conns []net.Conn

	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			if len(conns) == 0 {
				conn, err := net.Dial("tcp", ip+":8440")
				if err != nil {
					fmt.Println(err)
					continue
				}
				conns = append(conns, conn)
			}
		}
	}()
redo:
	cmd := exec.Command("tcpdump", "-n")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("failed pip stdout", err)
		time.Sleep(2 * time.Second)
		goto redo
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("failed start ", err)
		time.Sleep(2 * time.Second)
		goto redo
	}
	var buffError bytes.Buffer
	cmd.Stderr = &buffError

	buf := new(bytes.Buffer)
	lineBreak := []byte("\n")
	c := make(chan int, 1)
	scan := bufio.NewScanner(stdout)
	go func() {
		for scan.Scan() {
			if err := scan.Err(); err != nil {
				fmt.Println("failed to scan")
				break
			}
			vals := scan.Bytes()
			if len(conns) > 0 {
				buf.Write(vals)
				buf.Write(lineBreak)
				_, err := conns[0].Write(buf.Bytes())
				if err != nil {
					fmt.Println("failed to write", err)
					conns = make([]net.Conn, 0)
				}
				buf.Reset()
			} else {
				fmt.Printf("data: %s\n", vals)
			}
			//fields := strings.Fields(vals)
			//for _, field := range fields {
			//}
			select {
			case <-c:
				return
			default:
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Println("failed to start exec", err, buffError.String())
		c <- 1
		time.Sleep(2 * time.Second)
		goto redo
	}
}
