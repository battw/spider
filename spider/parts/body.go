package parts

import (
	"log"
)

type CtrlMsg struct {
	Kind string
	Body interface{}
}

type body struct {
	ctrlChan chan CtrlMsg
}

func Body() chan<- CtrlMsg {
	b := &body{make(chan CtrlMsg)}
	go func() {
		for cm := range b.ctrlChan {
			b.log(cm)
		}
	}()
	return b.ctrlChan
}

func (b *body) AttachLeg(l *leg) {

}

func (b *body) log(text interface{}) {
	log.Printf("Body: %v\n", text)
}
