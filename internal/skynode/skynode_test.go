// Copyright © 2018 BigOokie
//
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
package skynode

import (
	"testing"

	"github.com/go-test/deep"
)

func Test_NodeInfoString(t *testing.T) {
	nodeInfo := NodeInfo{
		Key:         "02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd",
		Conntype:    "TCP",
		SendBytes:   1,
		RecvBytes:   2,
		LastAckTime: 3,
		StartTime:   4}

	expectStr := "Key: 02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd, " +
		"Type: TCP, " +
		"SendBytes: 1, " +
		"RecvBytes: 2, " +
		"LastAckTime: 3s, " +
		"StartTime: 4s"

	resultStr := nodeInfo.String()

	if resultStr != expectStr {
		t.Errorf("Result:\n%s\nExpected:%s\n", resultStr, expectStr)
	}
}

func Test_NodeInfoFmtString(t *testing.T) {
	nodeInfo := NodeInfo{
		Key:         "02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd",
		Conntype:    "TCP",
		SendBytes:   1,
		RecvBytes:   2,
		LastAckTime: 3,
		StartTime:   4}

	expectStr := "Node Information:\n" +
		"Key        : 02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd\n" +
		"Type       : TCP\n" +
		"SendBytes  : 1\n" +
		"RecvBytes  : 2\n" +
		"LastAckTime: 3s\n" +
		"StartTime  : 4s"

	resultStr := nodeInfo.FmtString()

	if resultStr != expectStr {
		t.Errorf("Result:\n%s\nExpected:%s\n", resultStr, expectStr)
	}
}

func Test_NodesAreEqual_ExactOK(t *testing.T) {
	nodeA := NodeInfo{
		Key:         "02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd",
		Conntype:    "TCP",
		SendBytes:   1,
		RecvBytes:   2,
		LastAckTime: 3,
		StartTime:   4}

	nodeB := NodeInfo{
		Key:         "02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd",
		Conntype:    "TCP",
		SendBytes:   1,
		RecvBytes:   2,
		LastAckTime: 3,
		StartTime:   4}

	if !NodesAreEqual(nodeA, nodeB) {
		t.Fail()
	}

	if diff := deep.Equal(nodeA, nodeB); diff != nil {
		t.Error(diff)
	}
}

func Test_NodesAreEqual_KeysOnlyOK(t *testing.T) {
	nodeA := NodeInfo{
		Key:         "02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd",
		Conntype:    "TCP",
		SendBytes:   1,
		RecvBytes:   2,
		LastAckTime: 3,
		StartTime:   4}

	nodeB := NodeInfo{
		Key:         "02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd",
		Conntype:    "",
		SendBytes:   0,
		RecvBytes:   0,
		LastAckTime: 0,
		StartTime:   0}

	if !NodesAreEqual(nodeA, nodeB) {
		t.Fail()
	}

	if diff := deep.Equal(nodeA, nodeB); diff == nil {
		t.Fail()
	}
}

func Test_NodesAreEqual_Fail(t *testing.T) {
	nodeA := NodeInfo{
		Key:         "02b9d1cab7467771ce2bc8fd7c7340bba0c2a511004650064bcb368386263694fd",
		Conntype:    "TCP",
		SendBytes:   1,
		RecvBytes:   2,
		LastAckTime: 3,
		StartTime:   4}

	nodeB := NodeInfo{
		Key:         "ABC",
		Conntype:    "",
		SendBytes:   0,
		RecvBytes:   0,
		LastAckTime: 0,
		StartTime:   0}

	if NodesAreEqual(nodeA, nodeB) {
		t.Fail()
	}
}

func Test_NodeInfoSliceToMap(t *testing.T) {
	nodeA := NodeInfo{
		Key:         "NODE1KEY",
		Conntype:    "TCP",
		SendBytes:   1,
		RecvBytes:   2,
		LastAckTime: 3,
		StartTime:   4}

	nodeB := NodeInfo{
		Key:         "NODE2KEY",
		Conntype:    "TCP",
		SendBytes:   4,
		RecvBytes:   3,
		LastAckTime: 2,
		StartTime:   1}

	nodeSlice := NodeInfoSlice{
		nodeA,
		nodeB,
	}

	nodeMap := NodeInfoSliceToMap(nodeSlice)

	if len(nodeMap) != 2 {
		t.Errorf("Expected length 2, was %v", len(nodeMap))
	}

	for _, node := range nodeSlice {
		_, hasKey := nodeMap[node.Key]
		if !hasKey {
			t.Errorf("Node [%s] not found in NodeMap", node.Key)
		}

		if diff := deep.Equal(node, nodeMap[node.Key]); diff != nil {
			t.Error(diff)
		}
	}

}
