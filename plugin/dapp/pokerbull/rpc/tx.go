// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

type PokerBullStartTx struct {
	Value     int64 `json:"value"`
	PlayerNum int32 `json:"playerNum"`
	Fee       int64 `json:"fee"`
}

type PBContinueTxReq struct {
	GameId string `json:"gameId"`
	Fee    int64  `json:"fee"`
}

type PBQuitTxReq struct {
	GameId string `json:"gameId"`
	Fee    int64  `json:"fee"`
}

type PBQueryReq struct {
	GameId string `json:"GameId"`
	Fee    int64  `json:"fee"`
}
