package timer

import (
	dm "library/core/datamsg"
	"library/idgen"
	"library/logger"
	"service"
)

const ServiceName = "timer"

const (
	Break = iota
	Continue
	Return
)

type timerType struct {
	service.Service
	Output    *dm.DataMsgPipe
	timerPool map[idgen.ObjectID]*dm.DataMsg
}

func NewTimer() *timerType {
	t := &timerType{}
	t.Service = *service.NewService(ServiceName)
	t.Output = nil
	t.timerPool = make(map[idgen.ObjectID]*dm.DataMsg)
	return t
}

func (t *timerType) timer() {
	logger.Info("%s:service running", t.Name)
	var next, fun int = Continue, service.FunUnknown
	for {
		select {
		case msg, ok := <-t.Cmd:
			if !ok {
				logger.Info("Cmd Read error")
				break
			}
			next, fun = t.controlEntry(msg)
			break
		case msg, ok := <-t.ReadPipe():
			if !ok {
				logger.Info("Data Read error")
				break
			}
			next, fun = t.dataEntry(msg)
			if fun == service.FunDataPipeFull {
				logger.Warn("need do something when full")
			}
			break
		}

		switch next {
		case Break:
			break
		case Return:
			return
		case Continue:
		}
	}
	return

}
