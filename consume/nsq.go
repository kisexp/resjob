package main

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
	core "twist/core/json"
)

type PubInfo struct {
	Done     chan struct{}
	MsgBody  []byte
	StartPub time.Time
	Err      error
}

type PubInfoChan chan *PubInfo
type MPubInfoChan chan *MPubInfo

type Topic struct {
	sync.Mutex
	tname           string
	fullName        string
	pubWaitingChan  PubInfoChan
	mpubWaitingChan MPubInfoChan
	quitChan        chan struct{}
	pubFailedCnt    int64
}

func (t *Topic) GetWaitChan() PubInfoChan {
	return t.pubWaitingChan
}

func (t *Topic) QuitChan() <-chan struct{} {
	return t.quitChan
}

func (t *Topic) GetFullName() string {
	return t.fullName
}

func (t *Topic) PubFailed() int64 {
	return atomic.LoadInt64(&t.pubFailedCnt)
}

func (t *Topic) IncrPubFailed() {
	atomic.AddInt64(&t.pubFailedCnt, 1)
}

func (t *Topic) GetTopicName() string {
	return t.tname
}

func (t *Topic) GetMWaitChan() MPubInfoChan {
	return t.mpubWaitingChan
}

var serverPubFailedCnt int64

func incrServerPubFailed() {
	atomic.AddInt64(&serverPubFailedCnt, 1)
}

func internalPubAsync(clientTimer *time.Timer, msgBody []byte, topic *Topic) error {
	info := &PubInfo{
		Done:     make(chan struct{}),
		MsgBody:  msgBody,
		StartPub: time.Now(),
	}

	select {
	case topic.GetWaitChan() <- info:
	default:
		//if clientTimer == nil {
		//	clientTimer = time.NewTimer(pubWaitTimeout)
		//} else {
		//	if !clientTimer.Stop() {
		//		select {
		//		case <-clientTimer.C:
		//		default:
		//		}
		//	}
		//	clientTimer.Reset(pubWaitTimeout)
		//}
		//defer clientTimer.Stop()
		//select {
		//case topic.GetWaitChan() <- info:
		//case <-topic.QuitChan():
		//	log.Printf("topic %v put messages failed at exiting", topic.GetFullName())
		//	return errors.New("exiting")
		//case <-clientTimer.C:
		//	log.Printf("topic %v put messages timeout ", topic.GetFullName())
		//	topic.IncrPubFailed()
		//	incrServerPubFailed()
		//	return errors.New("pub to wait channel timeout")
		//}
	}
	<-info.Done
	return info.Err
}

type MessageID uint64

type MPubInfo struct {
	Done     chan struct{}
	Msgs     []*Message
	StartPub time.Time
	Err      error
}

type Message struct {
	ID        MessageID
	TraceID   uint64
	Body      []byte
	Timestamp int64
	attempts  uint32
	ExtBytes  []byte
}

const (
	maxBatchNum    = 0
	pubWaitTimeout = time.Second * 3
)

var testPopQueueTimeout int32

func NewMessage(id MessageID, body []byte) *Message {
	return &Message{
		ID:        id,
		TraceID:   0,
		Body:      body,
		Timestamp: time.Now().UnixNano(),
	}
}

func internalPubLoop(topic *Topic) {
	messages := make([]*Message, 0, 100)
	pubInfoList := make([]*PubInfo, 0, 100)
	mpubInfoList := make([]*MPubInfo, 0, 100)
	topicName := topic.GetTopicName()
	log.Printf("start pub loop for topic: %v ", topic.GetFullName())
	defer func() {
		done := false
		for !done {
			select {
			case info := <-topic.GetWaitChan():
				pubInfoList = append(pubInfoList, info)
			case minfo := <-topic.GetMWaitChan():
				mpubInfoList = append(mpubInfoList, minfo)
			default:
				done = true
			}
		}
		log.Printf("quit pub loop for topic: %v, left: %v, %v ", topic.GetFullName(), len(pubInfoList), len(mpubInfoList))
		for _, info := range pubInfoList {
			info.Err = errors.New("exiting")
			close(info.Done)
		}
		for _, info := range mpubInfoList {
			info.Err = errors.New("exiting")
			close(info.Done)
		}
	}()
	//quitChan := topic.QuitChan()
	infoChan := topic.GetWaitChan()
	//minfoChan := topic.GetMWaitChan()
	for {
		if len(messages) > maxBatchNum {
			infoChan = nil
			//minfoChan = nil
		} else {
			infoChan = topic.GetWaitChan()
			//minfoChan = topic.GetMWaitChan()
		}
		select {
		//case <-quitChan:
		//	return
		//case minfo := <-minfoChan:
		//	if time.Since(minfo.StartPub) >= pubWaitTimeout || atomic.LoadInt32(&testPopQueueTimeout) == 1 {
		//		topic.IncrPubFailed()
		//		incrServerPubFailed()
		//		minfo.Err = errors.New("pub timeout while pop wait queue")
		//		close(minfo.Done)
		//		log.Printf("topic %v put message timeout while pop queue, pub start: %s", topic.GetFullName(), minfo.StartPub)
		//		continue
		//	}
		//	messages = append(messages, minfo.Msgs...)
		//	mpubInfoList = append(mpubInfoList, minfo)
		case info := <-infoChan:
			if len(info.MsgBody) <= 0 {
				log.Println("empty msg body")
			}
			if time.Since(info.StartPub) >= pubWaitTimeout || atomic.LoadInt32(&testPopQueueTimeout) == 1 {
				topic.IncrPubFailed()
				incrServerPubFailed()
				info.Err = errors.New("pub timeout while pop wait queue")
				close(info.Done)
				log.Printf("topic %v put message timeout while pop queue, pub start: %s", topic.GetFullName(), info.StartPub)
				continue
			}
			messages = append(messages, NewMessage(0, info.MsgBody))
			pubInfoList = append(pubInfoList, info)
		default:
			if len(messages) == 0 {
				select {
				//case <-quitChan:
				//	return
				case info := <-infoChan:
					if time.Since(info.StartPub) >= pubWaitTimeout || atomic.LoadInt32(&testPopQueueTimeout) == 1 {
						topic.IncrPubFailed()
						incrServerPubFailed()
						info.Err = errors.New("pub timeout while pop wait queue")
						close(info.Done)
						log.Printf("topic %v put message timeout while pop queue, pub start: %s", topic.GetFullName(), info.StartPub)
						continue
					}
					messages = append(messages, NewMessage(0, info.MsgBody))
					pubInfoList = append(pubInfoList, info)
					//case minfo := <-minfoChan:
					//	if time.Since(minfo.StartPub) >= pubWaitTimeout || atomic.LoadInt32(&testPopQueueTimeout) == 1 {
					//		topic.IncrPubFailed()
					//		incrServerPubFailed()
					//		minfo.Err = errors.New("pub timeout while pop wait queue")
					//		close(minfo.Done)
					//		continue
					//	}
					//	messages = append(messages, minfo.Msgs...)
					//	mpubInfoList = append(mpubInfoList, minfo)
				}
				continue
			}
			if tcnt := atomic.LoadInt32(&testPopQueueTimeout); tcnt >= 1 {
				time.Sleep(time.Second * time.Duration(tcnt))
			}
			log.Printf("%s success import db", topicName)
			for _, info := range pubInfoList {
				info.Err = nil
				close(info.Done)
			}
			for _, minfo := range mpubInfoList {
				minfo.Err = nil
				close(minfo.Done)
			}
			pubInfoList = pubInfoList[:0]
			mpubInfoList = mpubInfoList[:0]
			messages = messages[:0]
		}

	}
}

func main() {
	topicCh := make(chan *Topic, 0)

	go func() {
		for {
			select {
			case topic := <-topicCh:
				//item := map[string]interface{}{
				//	"msgId":   i,
				//	"content": fmt.Sprintf("hello world->%d", i),
				//}
				//fmt.Println(topic)
				item := map[string]interface{}{
					"content": "hello world",
				}
				msgBody, err := core.JSON.Marshal(item)
				if err != nil {
					log.Printf("Marshal err: %v", err)
					continue
				}
				go func() {
					for {
						internalPubLoop(topic)
					}

				}()

				err = internalPubAsync(nil, msgBody, topic)
				if err != nil {
					log.Printf("internalPubAsync err: %v", err)
					continue
				}

			default:


			}
		}
	}()

	topic := &Topic{
		tname:           "tname",
		fullName:        "name-full",
		pubWaitingChan:  make(PubInfoChan, 0),
		mpubWaitingChan: make(MPubInfoChan, 0),
		quitChan:        make(chan struct{}, 0),
		pubFailedCnt:    0,
	}
	for i := 1; i <= 1000; i++ {
		topicCh <- topic

	}

}
