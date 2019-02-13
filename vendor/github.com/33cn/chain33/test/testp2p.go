package test

import (
"fmt"
"sync/atomic"
"time"

l "github.com/33cn/chain33/common/log/log15"
"github.com/33cn/chain33/queue"
"github.com/33cn/chain33/types"

// register gzip
_ "google.golang.org/grpc/encoding/gzip"
)

var (
	log = l.New("module", "testp2p")
)

// P2p interface
type TestP2p struct {
	client       queue.Client
	total        int64
	mode         int32  //0 只发 1 发后等待响应
	replyInterval  []int64
}

// New produce a p2p object
func New() *TestP2p {

	p2p := new(TestP2p)
	p2p.total = 10000
	p2p.mode = 1
	p2p.replyInterval = make([]int64, 0, p2p.total)

	p2p.node = node
	p2p.p2pCli = NewP2PCli(p2p)
	p2p.txFactory = make(chan struct{}, 1000)    // 1000 task
	p2p.otherFactory = make(chan struct{}, 1000) //other task 1000
	p2p.txCapcity = 1000
	return p2p
}

//Wait wait for ready
func (network *TestP2p) Wait() {}

func (network *TestP2p) isClose() bool {
	//return atomic.LoadInt32(&network.closed) == 1
	return false
}

// Close network client
func (network *P2p) Close() {
	atomic.StoreInt32(&network.closed, 1)
	log.Debug("close", "network", "ShowTaskCapcity done")

	if network.client != nil {
		network.client.Close()
	}
}

// SetQueueClient set the queue
func (network *P2p) SetQueueClient(client queue.Client) {
	network.client = client
	go func() {
		log.Info("p2p", "setqueuecliet", "ok")
		network.subTestP2pMsg()
	}()
}

func (network *P2p) showStatInfo() {
	ticker := time.NewTicker(time.Second * 15)
	log.Info("showStatInfo")
	defer ticker.Stop()
	for {
		if network.isClose() {
			log.Debug("showStatInfo", "loop", "done")
			return
		}
		<-ticker.C
		log.Debug("showStatInfo", "Capcity", atomic.LoadInt32(&network.txCapcity))
	}
}

func (network *P2p) subTestP2pMsg() {
	if network.client == nil {
		return
	}

	go network.showStatInfo()
	go func() {
		defer func() {
			close(network.otherFactory)
			close(network.txFactory)
		}()
		var taskIndex int64
		network.client.Sub("testp2p")
		for msg := range network.client.Recv() {
			if network.isClose() {
				log.Debug("subP2pMsg", "loop", "done")
				return
			}
			taskIndex++
			log.Debug("p2p recv", "msg", types.GetEventName(int(msg.Ty)), "msg type", msg.Ty, "taskIndex", taskIndex)
			if msg.Ty == types.EventTxBroadcast {
				network.txFactory <- struct{}{} //allocal task
				atomic.AddInt32(&network.txCapcity, -1)
			} else {
				if msg.Ty != types.EventPeerInfo {
					network.otherFactory <- struct{}{}
				}
			}
			switch msg.Ty {
			case types.EventTxBroadcast: //广播tx
				go network.p2pCli.BroadCastTx(msg, taskIndex)
			case types.EventBlockBroadcast: //广播block
				go network.p2pCli.BlockBroadcast(msg, taskIndex)
			case types.EventFetchBlocks:
				go network.p2pCli.GetBlocks(msg, taskIndex)
			case types.EventGetMempool:
				go network.p2pCli.GetMemPool(msg, taskIndex)
			case types.EventPeerInfo:
				go network.p2pCli.GetPeerInfo(msg, taskIndex)
			case types.EventFetchBlockHeaders:
				go network.p2pCli.GetHeaders(msg, taskIndex)
			case types.EventGetNetInfo:
				go network.p2pCli.GetNetInfo(msg, taskIndex)
			default:
				log.Warn("unknown msgtype", "msg", msg)
				msg.Reply(network.client.NewMessage("", msg.Ty, types.Reply{Msg: []byte("unknown msgtype")}))
				<-network.otherFactory
				continue
			}
		}
		log.Info("subP2pMsg", "loop", "close")

	}()

}

