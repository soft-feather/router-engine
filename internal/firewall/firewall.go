package firewall

import (
	"fmt"
	"github.com/asavie/xdp"
	"net"
)

func GenerationXskOnDev(devName string, rxIndex int) (xsk *xdp.Socket, err error) {
	dev, err2 := net.InterfaceByName(devName)
	if err2 != nil {
		return nil, err2
	}
	//name, err := netlink.LinkByName(devName)
	//if err != nil {
	//	return nil, err
	//}
	//queues := name.Attrs().NumRxQueues
	xsk, err = xdp.NewSocket(dev.Index, rxIndex, nil)
	if err != nil {
		fmt.Printf("error: failed to create an XDP socket: %v\n", err)
		return nil, err
	}
	return xsk, nil
}
