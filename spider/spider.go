package spider

import (
	"github.com/battw/spider/spider/parts"
	"github.com/gorilla/websocket"
	"log"
)

type Spider struct {
	ctrlChan chan<- parts.CtrlMsg
}

func Hatch() *Spider {
	return &Spider{parts.Body()}
}

func (s *Spider) GrowLeg(conn *websocket.Conn) {
	s.log("growing leg")
	testMsg := parts.CtrlMsg{Kind: "attach", Body: nil}
	s.ctrlChan <- testMsg
	//var lChan chan<- parts.Msg = parts.NewLeg(conn)
	//s.body.AttachLeg(l)
}

func (s *Spider) log(text interface{}) {
	log.Printf("Spider: %v\n", text)
}
