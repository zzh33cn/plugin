// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"testing"
)

func TestAddressBig(t *testing.T) {
	saddr := "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
	addr := StringToAddress(saddr)
	baddr := addr.Big()
	naddr := BigToAddress(baddr)
	if saddr != naddr.String() {
		t.Fail()
	}
}

func TestAddressBytes(t *testing.T) {
	addr := BytesToAddress([]byte{1})

	fmt.Println(addr.String())
}
