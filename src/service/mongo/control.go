package mongo

import (
	cm "library/core/controlmsg"
	dm "library/core/datamsg"
	"library/logger"
	"service"
)

func (t *mongoType) Start(bus *dm.DataMsgPipe) bool {
	logger.Info("%s:start running", t.Name)
	t.output = bus
	go t.mongo()
	return true
}

func (t *mongoType) Pause() bool {
	return true
}

func (t *mongoType) Resume() bool {
	return true
}

func (t *mongoType) Exit() bool {
	return true
}

func (t *mongoType) controlEntry(msg *cm.ControlMsg) (int, int) {
	switch msg.MsgType {
	case cm.ControlMsgExit:
		logger.Info("%s:ControlMsgPipe.Cmd Read %d", t.Name, msg.MsgType)
		t.Echo <- &cm.ControlMsg{MsgType: cm.ControlMsgExit}
		logger.Info("%s:exit", t.Name)
		return Return, service.FunOK
	case cm.ControlMsgPause:
		logger.Info("%s:paused", t.Name)
		t.Echo <- &cm.ControlMsg{MsgType: cm.ControlMsgPause}
		for {
			var resume bool = false
			select {
			case msg, ok := <-t.Cmd:
				if !ok {
					logger.Info("%s:Cmd Read error", t.Name)
					break
				}
				switch msg.MsgType {
				case cm.ControlMsgResume:
					t.Echo <- &cm.ControlMsg{MsgType: cm.ControlMsgResume}
					resume = true
					break
				}
			}
			if resume {
				break
			}
		}
		logger.Info("%s:resumed", t.Name)
	}
	return Continue, service.FunOK
}
