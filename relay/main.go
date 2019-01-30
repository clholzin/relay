package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	iface   = flag.String("i", "eth0", "Interface to read packets from")
	fname   = flag.String("r", "", "Filename to read from, overrides -i")
	snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
	tstype  = flag.String("timestamp_type", "", "Type of timestamps to use")
	promisc = flag.Bool("promisc", true, "Set promiscuous mode")

	printer     = flag.Bool("print", true, "Print out packets, if false only prints out statistics")
	maxcount    = flag.Int("c", -1, "Only grab this many packets, then exit")
	decoder     = flag.String("decoder", "Ethernet", "Name of the decoder to use")
	dump        = flag.Bool("X", false, "If true, dump very verbose info on each packet")
	statsevery  = flag.Int("stats", 1000, "Output statistics every N packets")
	printErrors = flag.Bool("errors", false, "Print out packet dumps of decode errors, useful for checking decoders against live traffic")
	lazy        = flag.Bool("lazy", true, "If true, do lazy decoding")
	defrag      = flag.Bool("defrag", true, "If true, do IPv4 defrag")
)

var (
	ip        = os.Getenv("hubip")
	conns     []net.Conn
	lineBreak = []byte("\n")
)

func main() {
	defer util.Run()()
	var handle *pcap.Handle
	var err error

	if *fname != "" {
		if handle, err = pcap.OpenOffline(*fname); err != nil {
			log.Fatal("PCAP OpenOffline error:", err)
		}
	} else {
		if handle, err = pcap.OpenLive(*iface, int32(*snaplen), *promisc, time.Duration(2*time.Second)); err != nil {
			log.Fatal("PCAP OpenLive error:", err)
		}
		defer handle.Close()
	}
	if len(flag.Args()) > 0 {
		bpffilter := strings.Join(flag.Args(), " ")
		fmt.Fprintf(os.Stderr, "Using BPF filter %q\n", bpffilter)
		if err = handle.SetBPFFilter(bpffilter); err != nil {
			log.Fatal("BPF filter error:", err)
		}
	}
	Connections()
	Run(handle)
}

func Connections() {
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

}

func Run(src gopacket.PacketDataSource) {
	if !flag.Parsed() {
		log.Fatalln("Run called without flags.Parse() being called")
	}
	var dec gopacket.Decoder
	var stamp time.Time
	var ok bool
	if dec, ok = gopacket.DecodersByLayerName[*decoder]; !ok {
		log.Fatalln("No decoder named", *decoder)
	}
	source := gopacket.NewPacketSource(src, dec)
	source.Lazy = *lazy
	source.NoCopy = true
	source.DecodeStreamsAsDatagrams = true
	fmt.Fprintln(os.Stderr, "Starting to read packets")
	count := 0
	bytes := int64(0)
	start := time.Now()

	errors := 0
	truncated := 0
	layertypes := map[gopacket.LayerType]int{}
	defragger := ip4defrag.NewIPv4Defragmenter()

	for packet := range source.Packets() {
		count++
		bytes += int64(len(packet.Data()))

		// defrag the IPv4 packet if required
		if *defrag {
			ip4Layer := packet.Layer(layers.LayerTypeIPv4)
			if ip4Layer == nil {
				continue
			}
			ip4 := ip4Layer.(*layers.IPv4)
			l := ip4.Length

			newip4, err := defragger.DefragIPv4(ip4)
			if err != nil {
				log.Fatalln("Error while de-fragmenting", err)
			} else if newip4 == nil {
				continue // packet fragment, we don't have whole packet yet.
			}
			if newip4.Length != l {
				fmt.Printf("Decoding re-assembled packet: %s\n", newip4.NextLayerType())
				pb, ok := packet.(gopacket.PacketBuilder)
				if !ok {
					panic("Not a PacketBuilder")
				}
				nextDecoder := newip4.NextLayerType()
				nextDecoder.Decode(newip4.Payload, pb)
			}
		}

		if len(conns) > 0 {
			l := packet.TransportLayer()
			if l == nil {
				continue
			}
			n := packet.NetworkLayer()
			if l == nil {
				continue
			}

			if meta := packet.Metadata(); meta != nil {
				stamp = meta.CaptureInfo.Timestamp
			} else {
				stamp = time.Now()
			}

			tLayer := fmt.Sprintf("%s", gopacket.LayerString(l))
			nLayer := fmt.Sprintf("%s", gopacket.LayerString(n))

			//2018-10-24 06:20:30.297118 +0000 UTC  // stamp
			//(32 bytes) // len(l.LayerContents()),

			//prevent recaptured packes
			if strings.Contains(tLayer, "=8440") {
				continue
			}

			packageParsed := parseLayer(nLayer, tLayer)
			packageParsed["TimeStamp"] = stamp
			packageParsed["Data"] = string(packet.Data())

			packetBytes, err := json.Marshal(packageParsed)
			if err != nil {
				fmt.Println("ERR: failed to marshal ", err)
			}
			packetBytes = append(packetBytes, lineBreak...)
			_, err = conns[0].Write(packetBytes)
			if err != nil {
				fmt.Println("failed to write", err)
				conns = make([]net.Conn, 0)
			}

		}
		if *printer {
			fmt.Printf("%+v\n", packet)
			//fmt.Printf("%s\n\n", packet.Data())

		} else if *dump {
			fmt.Println(packet.Dump())
		}

		if !*lazy || *printer || *dump { // if we've already decoded all layers...
			for _, layer := range packet.Layers() {
				layertypes[layer.LayerType()]++
			}
			if packet.Metadata().Truncated {
				truncated++
			}
			if errLayer := packet.ErrorLayer(); errLayer != nil {
				errors++
				if *printErrors {
					fmt.Println("Error:", errLayer.Error())
					fmt.Println("--- Packet ---")
					fmt.Println(packet.Dump())
				}
			}
		}
		done := *maxcount > 0 && count >= *maxcount
		if count%*statsevery == 0 || done {
			fmt.Fprintf(os.Stderr, "Processed %v packets (%v bytes) in %v, %v errors and %v truncated packets\n", count, bytes, time.Since(start), errors, truncated)
			if len(layertypes) > 0 {
				fmt.Fprintf(os.Stderr, "Layer types seen: %+v\n", layertypes)
			}
		}
		if done {
			break
		}
	}
}

//l2
//IPv4	{Contents=[..20..] Payload=[..38..] Version=4 IHL=5 TOS=0 Length=58 Id=1626 Flags=DF FragOffset=0 TTL=63 Protocol=TCP Checksum=56632 SrcIP=172.17.0.5 DstIP=172.17.0.4 Options=[] Padding=[]}
//l3
//TCP	{Contents=[..32..] Payload=[..6..] SrcPort=53126 DstPort=8222 Seq=540397798 Ack=1525309599 DataOffset=8 FIN=false SYN=false RST=false PSH=true ACK=true URG=false ECE=false CWR=false NS=false Window=229 Checksum=22616 Urgent=0 Options=[TCPOption(NOP:), TCPOption(NOP:), TCPOption(Timestamps:12277163/12277063 0x00bb55ab00bb5547)] Padding=[]}

func parseLayer(l2, l3 string) (u map[string]interface{}) {
	u = make(map[string]interface{})

	l2 = strings.Replace(l2, "{", "", -1)
	l2 = strings.Replace(l2, "}", "", -1)

	l3 = strings.Replace(l3, "{", "", -1)
	l3 = strings.Replace(l3, "}", "", -1)

	u = parseLayerFormat(l2, u)
	u = parseLayerFormat(l3, u)
	fmt.Println(u)

	return
}

func parseLayerFormat(v string, d map[string]interface{}) (g map[string]interface{}) {
	split := strings.Fields(v)
	for _, s := range split {
		if strings.Index(s, "=") > -1 {
			vals := strings.Split(s, "=")
			if len(vals) == 2 {
				key := vals[0]
				d[key] = vals[1]
			}
		}
	}
	g = d
	return
}
