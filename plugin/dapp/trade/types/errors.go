// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	ErrTSellBalanceNotEnough   = errors.New("ErrTradeSellBalanceNotEnough")
	ErrTSellOrderNotExist      = errors.New("ErrTradeSellOrderNotExist")
	ErrTSellOrderNotStart      = errors.New("ErrTradeSellOrderNotStart")
	ErrTSellOrderNotEnough     = errors.New("ErrTradeSellOrderNotEnough")
	ErrTSellOrderSoldout       = errors.New("ErrTradeSellOrderSoldout")
	ErrTSellOrderRevoked       = errors.New("ErrTradeSellOrderRevoked")
	ErrTSellOrderExpired       = errors.New("ErrTradeSellOrderExpired")
	ErrTSellOrderRevoke        = errors.New("ErrTradeSellOrderRevokeNotAllowed")
	ErrTSellNoSuchOrder        = errors.New("ErrTradeSellNoSuchOrder")
	ErrTBuyOrderNotExist       = errors.New("ErrTradeBuyOrderNotExist")
	ErrTBuyOrderNotEnough      = errors.New("ErrTradeBuyOrderNotEnough")
	ErrTBuyOrderSoldout        = errors.New("ErrTradeBuyOrderSoldout")
	ErrTBuyOrderRevoked        = errors.New("ErrTradeBuyOrderRevoked")
	ErrTBuyOrderRevoke         = errors.New("ErrTradeBuyOrderRevokeNotAllowed")
	ErrTCntLessThanMinBoardlot = errors.New("ErrTradeCountLessThanMinBoardlot")
)
