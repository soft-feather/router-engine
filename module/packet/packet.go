package packet

import (
	"fmt"
	"log"
	"net"
)

// Server
// packet 服务模块 , 用于处理全局的数据包 , 网络协议转换
// packet 意为 数据包
// 从二层包开始 , 一直到应用层的包 , 都是数据包
type Server struct {
}

func (p *Server) Init() error {

	return nil
}

func (p *Server) Shutdown() {

}

func Test() {
	// 创建Packet Socket
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 创建一个缓冲区来接收数据
	buf := make([]byte, 1024)

	for {
		// 读取数据
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
		}

		// 打印接收到的数据
		fmt.Printf("Received %d bytes from %s\n", n, addr)
		fmt.Println(buf[:n])
	}
}
