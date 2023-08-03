package firewall

import (
	"encoding/hex"
	"fmt"
	"github.com/aquasecurity/libbpfgo"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"os"
	"unsafe"
)

func main() {
	devName := "wlp0s20f3"
	bpfModule, err2 := libbpfgo.NewModuleFromFile("./main.bpf.o")
	if err2 != nil {
		panic(err2)
	}
	defer bpfModule.Close()
	err := bpfModule.BPFLoadObject()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	program, err2 := bpfModule.GetProgram("xdp_prog")
	if err2 != nil {
		panic(err2)
	}
	attachXDP, err2 := program.AttachXDP(devName)
	if err2 != nil {
		panic(err2)
	}
	defer attachXDP.Destroy()

	xsk, err := GenerationXskOnDev(devName, 0)
	if err != nil {
		panic(err)
	}

	index := 0
	xskMap, err2 := bpfModule.GetMap("xsk_maps")
	fd := xsk.FD()
	err = xskMap.Update(unsafe.Pointer(&index), unsafe.Pointer(&fd))
	if err != nil {
		fmt.Println(err)
		return
	}
	bpfMap, err2 := bpfModule.GetMap("index_stat")
	err = bpfMap.Update(unsafe.Pointer(&index), unsafe.Pointer(&index))
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		// If there are any free slots on the Fill queue...
		if n := xsk.NumFreeFillSlots(); n > 0 {
			// ...then fetch up to that number of not-in-use
			// descriptors and push them onto the Fill ring queue
			// for the kernel to fill them with the received
			// frames.
			xsk.Fill(xsk.GetDescs(n))
		}

		// Wait for receive - meaning the kernel has
		// produced one or more descriptors filled with a received
		// frame onto the Rx ring queue.
		log.Printf("waiting for frame(s) to be received...")
		numRx, _, err := xsk.Poll(-1)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}

		if numRx > 0 {
			// Consume the descriptors filled with received frames
			// from the Rx ring queue.
			rxDescs := xsk.Receive(numRx)

			// Print the received frames and also modify them
			// in-place replacing the destination MAC address with
			// broadcast address.
			for i := 0; i < len(rxDescs); i++ {
				pktData := xsk.GetFrame(rxDescs[i])
				pkt := gopacket.NewPacket(pktData, layers.LayerTypeEthernet, gopacket.Default)
				log.Printf("received frame:\n%s%+v", hex.Dump(pktData[:]), pkt)
			}
		}
	}
}
