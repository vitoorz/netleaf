package job

import (
//"time"
)

import (
	dm "library/core/datamsg"
	"library/logger"
	"service"
	"service/job/task"
	//ts "types/service"
	"types"
)

func (t *jobType) dataEntry(msg *dm.DataMsg) (operate int, funCode int) {
	defer func() {
		if x := recover(); x != nil {
			logger.Error("%s:job panic: %v", t.Name, x)
			logger.Stack()
		}
		operate = Continue
		funCode = service.FunPanic
	}()

	logger.Info("%s:get data msg:%d,payload:%v", t.Name, msg.MsgType, msg.Payload.([]byte))

	switch msg.MsgType {
	case types.MsgTypeTelnet:
		switch msg.Sender {
		case "mongo":
			msg.Receiver = "tcpserver"
		case "tcpserver":
			choosetask := task.Parse(string(msg.Payload.([]byte)))
			task.Route[choosetask](msg)
		}
	case types.MsgTypeUnknown:
		fallthrough
	default:
		logger.Warn("%s:not handle:get data msg from:%s,type:%d", t.Name, msg.Sender, msg.MsgType)
		msg.Receiver = dm.NoReceiver
	}
	msg.Sender = t.Name
	if msg.Receiver != dm.NoReceiver {
		ok := t.Output.WritePipeNB(msg)
		if !ok {
			// channel full
			return Continue, service.FunDataPipeFull
		}
	}
	return Continue, service.FunOK
}
