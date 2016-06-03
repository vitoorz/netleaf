package privatetcp

import (
	"io"
	"net"
)

import (
	cm "library/core/controlmsg"
	dm "library/core/datamsg"
	"library/logger"
	"service"
	"time"
	"types"
)

const ServiceName = "privatetcpserver"

type privateTCPServer struct {
	service.Service
	output *dm.DataMsgPipe

	listener *net.TCPListener
	ip       string
	port     string
}

func NewPrivateTCPServer(name, ip, port string) *privateTCPServer {
	t := &privateTCPServer{}
	t.Service = *service.NewService(ServiceName)
	t.output = nil
	t.Name = name
	t.ip = ip
	t.port = port
	t.Buffer = service.NewBufferPool(&t.Service)
	t.Instance = t
	return t
}

func (t *privateTCPServer) Start(bus *dm.DataMsgPipe) bool {
	logger.Info("%s:Start privateTCPServer", t.Name)
	t.output = bus
	tcpAddr, err := net.ResolveTCPAddr("tcp", t.ip+":"+t.port)
	if err != nil {
		logger.Error("%s:net.ResolveTCPAddr error,%s", t.Name, err.Error())
		return false
	}

	t.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logger.Error("%s:net.ListenTCP error,%s", t.Name, err.Error())
		return false
	}

	logger.Info("%s:listening port:%s", t.Name, t.port)
	go t.serve()
	return true
}

func (t *privateTCPServer) Pause() bool {
	return true
}

func (t *privateTCPServer) Resume() bool {
	return true
}

func (t *privateTCPServer) Exit() bool {
	return true
}

func (t *privateTCPServer) serve() {

	for {
		connect, err := t.listener.AcceptTCP()
		if err != nil {
			logger.Error("%s:listener.AcceptTCP error,%s", t.Name, err.Error())
			time.Sleep(time.Second * 2)
			connect.Close()
			continue
		}
		go t.readConn(connect)
	}
}

func (t *privateTCPServer) readConn(connection *net.TCPConn) {

	for {
		var stream []byte
		for {
			data := make([]byte, 1)
			n, err := io.ReadAtLeast(connection, data, 1)
			if err != nil {
				logger.Warn("%s:read byte:%d,error:%s", t.Name, n, err.Error())
				connection.Close()
				return
			}
			if data[0] == 10 {
				break
			} else {
				stream = append(stream, data...)
			}
		}
		var d *dm.DataMsg
		d = dm.NewDataMsg("tcpserver", "job", types.MsgTypeTelnet, stream)
		d.SetMeta(t.Name, connection)
		t.output.WritePipeNB(d)
	}
}

func (t *privateTCPServer) ControlHandler(msg *cm.ControlMsg) (int, int) {
	return 0, 0
}

func (t *privateTCPServer) DataHandler(msg *dm.DataMsg) bool {
	meta, ok := msg.Meta(t.Name)
	if !ok {
		logger.Error("%s:wrong meta in datamsg:%+v", t.Name, msg)
		return false
	}
	connection := meta.(*net.TCPConn)
	//todo: need to verify if the data payload is []byte
	count, err := connection.Write(msg.Payload.([]byte))
	if err != nil {
		logger.Warn("%s:conn write err:%s", t.Name, err.Error())
		connection.Close()
		return false
	} else {
		logger.Info("%s:sent to network:%d byte", t.Name, count)
	}
	return true
}
