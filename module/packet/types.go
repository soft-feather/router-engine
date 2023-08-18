package packet

type LightStatus uint32

// 目前只有四种状态
// 初始状态: 初始化状态, 准备进入接收数据
// 可读状态: 当套接字变为“可读”状态时，这意味着有数据已经到达并且可以从套接字中读取。对于TCP，它可能意味着远程方已经发送了数据并且数据已经到达了本地系统。对于UDP，这可能意味着一个数据报已经到达
// 可写状态: 当套接字变为“可写”状态时，这意味着应用程序可以向其发送数据，而不会被阻塞。对于TCP，这通常意味着连接已经建立并且有足够的窗口大小可以发送更多的数据。需要注意的是，一个TCP套接字在大多数情况下都是可写的，除非它的发送缓冲区已满或接收方的接收窗口已满
// 关闭状态: 当套接字变为“关闭”状态时，这意味着套接字已经关闭。这意味着应用程序不能再从套接字中读取或写入数据。对于TCP，这意味着连接已经关闭，因为远程方已经发送了一个FIN数据包并且已经收到了一个ACK。对于UDP，这意味着套接字已经关闭，因为远程方已经发送了一个ICMP端口不可达消息
const (
	InitStatus = iota
	ReadableStatus
	WritableStatus
	ClosedStatus
)

// StatusChangeOptions
// 状态改变的选项
// 用于传递一些额外的信息, 扩展功能, 如指示如果发生错误, 是否需要关闭连接
type StatusChangeOptions struct {
}

// StatusChange
// 状态改变
type StatusChange struct {
	OldStatus LightStatus
	NewStatus LightStatus
	Options   StatusChangeOptions
}

// Changeable
// 可改变状态的对象, 被注册事件
type Changeable interface {
	Change(change StatusChange) error
	Read() []byte
	Write([]byte)
	init()
	close()
}

type LightIdBot struct {
	// focus id 用于标记当前的 id 最大值/最小值, 因为要么是递增, 要么是递减的 id 发放
	// focus id 目前使用递增算法, 所以 focus id 为最大值
	focusId uint32
	bot     []uint32
}

func (l *LightIdBot) init() {
	l.focusId = 100
	ls := make([]uint32, 100)
	for i := uint32(0); i < l.focusId; i++ {
		ls[i] = i
	}
	l.bot = ls
}

func (l *LightIdBot) pop() uint32 {
	if len(l.bot) == 0 {
		l.increase([]uint32{})
	}
	id := l.bot[0]
	l.bot = l.bot[1:]
	return id
}

// increase
// bot 增长, 一般是在 bot 用完的时候调用, 用于增加 bot 的容量, 要将 focus id 一并修改
// idLs 为 bot 的增长值, 如果不传入, 则默认增长 100
func (l *LightIdBot) increase(idLs []uint32) {
	if len(idLs) == 0 {
		l.focusId += 100
		idLs = make([]uint32, 100)
		for i := uint32(0); i < 100; i++ {
			idLs[i] = i + l.focusId
		}
	} else {
		l.bot = append(l.bot, idLs...)
		// TODO
	}

}

// GetLightId
// 获取一个可用的 light id, 机制是只要调用此方法一定会返回一个可用的 light id, 目前使用递增算法, 即每次返回的 light id 都会比上次大
func (l *LightIdBot) GetLightId() uint32 {
	return l.pop()
}

func GetLightId() uint32 {

	return 0
}

type Light struct {
	Id     uint32
	Status LightStatus
	buf    []byte
}

// init
// 初始化
func (l *Light) init() {
	l.Status = InitStatus
	l.buf = []byte{}
}

// Change
// TODO
func (l *Light) Change(change StatusChange) error {
	if change.OldStatus != l.Status {
		return ErrStatusNotMatch
	}
	l.Status = change.NewStatus
	return nil
}

// Read
// 将 buffer 中的数据读取出来
func (l *Light) Read() []byte {
	return l.buf
}

// Write
// 将数据写入 buffer
func (l *Light) Write(buf []byte) {
	l.buf = append(l.buf, buf...)
}

// Close
// 关闭连接
// TODO
func (l *Light) close() {
	l.Status = ClosedStatus
	l.buf = []byte{}
}
