package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type node struct {
	Index int    `json:"-"`
	Name  string `json:"name"`
}

type link struct {
	Source int `json:"source"`
	Target int `json:"target"`
	Value  int `json:"value"`
}

type Nodes []*node
type Links []*link

type DataModel struct {
	Nodes Nodes `json:"nodes"`
	Links Links `json:"links"`
}

var (
	ip           = os.Getenv("hostip")
	dir          = os.Getenv("dir")
	upgrader     = websocket.Upgrader{}
	dil          byte
	dilArray     []byte
	transmitData bytes.Buffer
	Data         = new(DataModel)
)

func init() {
	dilArray = append(dilArray, []byte("\n")...)
	dil = dilArray[0]
	fmt.Println(dilArray, dil)
	Data.Nodes = make([]*node, 0)
	Data.Links = make([]*link, 0)
}

func main() {
	fmt.Println("Starting Listener")
	listener, _ := net.Listen("tcp", ip+":8440")
	var conns []net.Conn
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println(err)
				time.Sleep(time.Second)
				continue
			}
			conns = append(conns, conn)
			go func(c net.Conn) {
				buf := bufio.NewScanner(c)
				for buf.Scan() {
					if err := buf.Err(); err != nil {
						fmt.Println(err)
						conns = make([]net.Conn, 0)
						return
					}
					data := buf.Bytes()
					fmt.Printf("%s\n", data)
					transmitData.Write(data)
					transmitData.Write(dilArray)
				}

			}(conn)
		}
	}()
	go resetCounts()
	go processData()
	http.HandleFunc("/", index)
	http.HandleFunc("/data", transmit)
	fmt.Println("Starting Server on 9999")
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func resetCounts() {
	tick := time.NewTicker(10 * time.Second)
	for range tick.C {
		for _, d := range Data.Links {
			d.Value = 1
		}
	}
}

func processData() {
	for {

		if transmitData.Len() == 0 {
			time.Sleep(100 * time.Millisecond)
			continue

		}
		sdata, err := transmitData.ReadBytes(dil)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		}
		if len(sdata) == 0 {
			continue
		}
		data := make(map[string]string)
		err = json.Unmarshal(sdata, &data)
		if err != nil {
			fmt.Println("Err: failed to UnMarshal data", err)
			continue
		}
		if len(data) > 0 {
			source := data["SrcIP"] + ":" + data["SrcPort"]
			target := data["DstIP"] + ":" + data["DstPort"]
			if !strings.Contains(source, ".") || len(source) <= 10 {
				continue
			}
			//source, target = setNames(source, target)
			var founds bool
			//var foundt bool
			var nodeIndex int
			//var nodeTindex int

			for i, d := range Data.Nodes {
				if d.Name == source { //source if does not exhist, make them
					nodeIndex = i
					founds = true
					break
				}
			}
			if founds {
				for _, d := range Data.Links {
					if d.Source == nodeIndex {
						d.Value++
					}
				}
			} else {
				nodeIndex = len(Data.Nodes)
				NewNode := &node{Index: nodeIndex, Name: source}
				Data.Nodes = append(Data.Nodes, NewNode)
				var targetIndex int
				for i, d := range Data.Nodes {
					if d.Name == target { //source if does not exhist, make them
						targetIndex = i
						founds = true
					}
				}
				if !founds {
					targetIndex = len(Data.Nodes)
					targetNode := &node{Index: targetIndex, Name: target}
					Data.Nodes = append(Data.Nodes, targetNode)
				}

				targetSourceIndex := len(Data.Nodes)
				NewNode2 := &node{Index: targetSourceIndex, Name: source}
				Data.Nodes = append(Data.Nodes, NewNode2)

				NewLink := &link{nodeIndex, targetIndex, 1}
				Data.Links = append(Data.Links, NewLink)

				TargetLink := &link{targetIndex, targetSourceIndex, 1}
				Data.Links = append(Data.Links, TargetLink)
			}
		}

	}

}

// example to match ip and set names
func setNames(source, target string) (string, string) {

	if strings.Contains(source, "172.17.0.2") {
		source = "Relay"
	} else if strings.Contains(source, "172.17.0.6") {
		source = "Relay Hub"
	} else if strings.Contains(source, "172.17.0.3") {
		source = "Proxy"
	} else if strings.Contains(source, "172.17.0.4") {
		source = "Aquirer2"
	} else if strings.Contains(source, "172.17.0.5") {
		source = "Aquirer"
	}
	if strings.Contains(target, "172.17.0.2") {
		target = "Relay"
	} else if strings.Contains(target, "172.17.0.6") {
		target = "Relay Hub"
	} else if strings.Contains(target, "172.17.0.3") {
		target = "Proxy"
	} else if strings.Contains(target, "172.17.0.4") {
		target = "Aquirer2"
	} else if strings.Contains(target, "172.17.0.5") {
		target = "Aquirer"
	}
	return source, target
}

func index(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(r.URL.Path)
	fmt.Println(path)
	if path == "/" || path == "." {
		path = "/index.html"
	}

	Data.Nodes = make([]*node, 0)
	Data.Links = make([]*link, 0)

	location := dir + path
	fmt.Println(location)
	file, err := os.Open(location)
	if err != nil {
		fmt.Println(err)
		return
	}
	indexData, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = w.Write(indexData)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func transmit(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		fmt.Fprintf(w, "failed")
		return
	}
	defer c.Close()

	for {
		data, err := json.Marshal(Data)
		if err != nil {
			fmt.Println("unmarshel err", err)
			time.Sleep(time.Second)
			continue
		}
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(10 * time.Second)
	}
}

// data generator
//      go func() {
//
//		ti := time.NewTicker(time.Second)
//		for range ti.C {
//			//	transmitData.Write([]byte(`{"data":"`))
//			transmitData.Write([]byte(`21:18:21.440179 IP 192.168.0.2.55615 > 13.57.54.63.443: Flags [P.], seq 273:318, ack 282, win 4096, options [nop,nop,TS val 691879038 ecr 185528752], length 45`))
//			//transmitData.Write([]byte(`"}`))
//			transmitData.Write(dilArray)
//			//	transmitData.Write([]byte(`{"data":"`))
//			transmitData.Write([]byte(`21:18:21.620072 IP 13.57.54.63.443 > 192.168.0.2.55615: Flags [.], ack 363, win 1261, options [nop,nop,TS val 185528782 ecr 691879038], length 0`))
//			//	transmitData.Write([]byte(`"}`))
//			transmitData.Write(dilArray)
//		}
//
//	}()
