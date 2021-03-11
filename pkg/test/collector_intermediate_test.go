// Copyright 2020 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build integration

package test

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/vmware/go-ipfix/pkg/collector"
	"github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/intermediate"
	"github.com/vmware/go-ipfix/pkg/registry"
)

// Run TestSingleRecordTCPTransport and TestSingleRecordTCPTransportIPv6 along with
// debug log for the message in pkg/exporter/process.go before sending it to get following
// raw bytes for template and data packets.
// Following data packets are generated with getTestRecord in exporter_collector_test.go
// dataPacket1IPv4: getTestRecord(false, false)
// dataPacket2IPv4: getTestRecord(true, false)
// dataPacket1IPv6: getTestRecord(false, true)
// dataPacket2IPv6: getTestRecord(true, true)
var templatePacketIPv4 = []byte{0x0, 0xa, 0x0, 0x70, 0x60, 0x48, 0x12, 0x4b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x2, 0x0, 0x60, 0x1, 0x0, 0x0, 0xf, 0x0, 0x8, 0x0, 0x4, 0x0, 0xc, 0x0, 0x4, 0x0, 0x7, 0x0, 0x2, 0x0, 0xb, 0x0, 0x2, 0x0, 0x4, 0x0, 0x1, 0x0, 0x97, 0x0, 0x4, 0x0, 0x56, 0x0, 0x8, 0x0, 0x2, 0x0, 0x8, 0x80, 0x65, 0xff, 0xff, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x67, 0xff, 0xff, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x6c, 0x0, 0x2, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x89, 0x0, 0x1, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x6a, 0x0, 0x4, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x56, 0x0, 0x8, 0x0, 0x0, 0x72, 0x79, 0x80, 0x2, 0x0, 0x8, 0x0, 0x0, 0x72, 0x79}
var dataPacket1IPv4 = []byte{0x0, 0xa, 0x0, 0x52, 0x60, 0x48, 0x12, 0x4b, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x42, 0xa, 0x0, 0x0, 0x1, 0xa, 0x0, 0x0, 0x2, 0x4, 0xd2, 0x16, 0x2e, 0x6, 0x4a, 0xf9, 0xf0, 0x70, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0xe8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xf4, 0x0, 0x4, 0x70, 0x6f, 0x64, 0x32, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x90, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xc8}
var dataPacket2IPv4 = []byte{0x0, 0xa, 0x0, 0x52, 0x60, 0x48, 0x63, 0xc8, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x42, 0xa, 0x0, 0x0, 0x1, 0xa, 0x0, 0x0, 0x2, 0x4, 0xd2, 0x16, 0x2e, 0x6, 0x4a, 0xf9, 0xf8, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xf4, 0x4, 0x70, 0x6f, 0x64, 0x31, 0x0, 0x12, 0x83, 0x2, 0xa, 0x0, 0x0, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x2c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x96}
var templatePacketIPv6 = []byte{0x0, 0xa, 0x0, 0x70, 0x60, 0x48, 0x12, 0x4b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x2, 0x0, 0x60, 0x1, 0x0, 0x0, 0xf, 0x0, 0x1b, 0x0, 0x10, 0x0, 0x1c, 0x0, 0x10, 0x0, 0x7, 0x0, 0x2, 0x0, 0xb, 0x0, 0x2, 0x0, 0x4, 0x0, 0x1, 0x0, 0x97, 0x0, 0x4, 0x0, 0x56, 0x0, 0x8, 0x0, 0x2, 0x0, 0x8, 0x80, 0x65, 0xff, 0xff, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x67, 0xff, 0xff, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x6c, 0x0, 0x2, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x89, 0x0, 0x1, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x6b, 0x0, 0x10, 0x0, 0x0, 0xdc, 0xba, 0x80, 0x56, 0x0, 0x8, 0x0, 0x0, 0x72, 0x79, 0x80, 0x2, 0x0, 0x8, 0x0, 0x0, 0x72, 0x79}
var dataPacket1IPv6 = []byte{0x0, 0xa, 0x0, 0x76, 0x60, 0x48, 0x12, 0x4b, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x66, 0x20, 0x1, 0x0, 0x0, 0x32, 0x38, 0xdf, 0xe1, 0x0, 0x63, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xfb, 0x20, 0x1, 0x0, 0x0, 0x32, 0x38, 0xdf, 0xe1, 0x0, 0x63, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xfc, 0x4, 0xd2, 0x16, 0x2e, 0x6, 0x4a, 0xf9, 0xf0, 0x70, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0xe8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xf4, 0x0, 0x4, 0x70, 0x6f, 0x64, 0x32, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x90, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xc8}
var dataPacket2IPv6 = []byte{0x0, 0xa, 0x0, 0x76, 0x60, 0x48, 0x63, 0xc8, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x1, 0x0, 0x0, 0x66, 0x20, 0x1, 0x0, 0x0, 0x32, 0x38, 0xdf, 0xe1, 0x0, 0x63, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xfb, 0x20, 0x1, 0x0, 0x0, 0x32, 0x38, 0xdf, 0xe1, 0x0, 0x63, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xfc, 0x4, 0xd2, 0x16, 0x2e, 0x6, 0x4a, 0xf9, 0xf8, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xf4, 0x4, 0x70, 0x6f, 0x64, 0x31, 0x0, 0x12, 0x83, 0x2, 0x20, 0x1, 0x0, 0x0, 0x32, 0x38, 0xbb, 0xbb, 0x0, 0x63, 0x0, 0x0, 0x0, 0x0, 0xaa, 0xaa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x2c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x96}

var (
	flowKeyRecordMap = make(map[intermediate.FlowKey]intermediate.AggregationFlowRecord)
	flowKey1         = intermediate.FlowKey{SourceAddress: "10.0.0.1", DestinationAddress: "10.0.0.2", Protocol: 6, SourcePort: 1234, DestinationPort: 5678}
	flowKey2         = intermediate.FlowKey{SourceAddress: "2001:0:3238:dfe1:63::fefb", DestinationAddress: "2001:0:3238:dfe1:63::fefc", Protocol: 6, SourcePort: 1234, DestinationPort: 5678}
	correlatefields  = []string{
		"sourcePodName",
		"sourcePodNamespace",
		"sourceNodeName",
		"destinationPodName",
		"destinationPodNamespace",
		"destinationNodeName",
		"destinationClusterIPv4",
		"destinationClusterIPv6",
		"destinationServicePort",
	}
	nonStatsElementList = []string{
		"flowEndSeconds",
	}
	statsElementList = []string{
		"packetTotalCount",
		"packetDeltaCount",
		"reversePacketTotalCount",
		"reversePacketDeltaCount",
	}
	antreaSourceStatsElementList = []string{
		"packetTotalCountFromSourceNode",
		"packetDeltaCountFromSourceNode",
		"reversePacketTotalCountFromSourceNode",
		"reversePacketDeltaCountFromSourceNode",
	}
	antreaDestinationStatsElementList = []string{
		"packetTotalCountFromDestinationNode",
		"packetDeltaCountFromDestinationNode",
		"reversePacketTotalCountFromDestinationNode",
		"reversePacketDeltaCountFromDestinationNode",
	}
	aggregationWorkerNum = 2
)

func init() {
	// Load the global registry
	registry.LoadRegistry()
}

func TestCollectorToIntermediateIPv4(t *testing.T) {
	address, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	testCollectorToIntermediate(t, address, false)
}

func TestCollectorToIntermediateIPv6(t *testing.T) {
	address, err := net.ResolveTCPAddr("tcp", "[::1]:0")
	if err != nil {
		t.Error(err)
	}
	testCollectorToIntermediate(t, address, true)
}

func testCollectorToIntermediate(t *testing.T, address net.Addr, isIPv6 bool) {
	aggregatedFields := &intermediate.AggregationElements{
		NonStatsElements:                   nonStatsElementList,
		StatsElements:                      statsElementList,
		AggregatedSourceStatsElements:      antreaSourceStatsElementList,
		AggregatedDestinationStatsElements: antreaDestinationStatsElementList,
	}
	// Initialize aggregation process and collecting process
	cpInput := collector.CollectorInput{
		Address:       address.String(),
		Protocol:      address.Network(),
		MaxBufferSize: 1024,
		TemplateTTL:   0,
		IsEncrypted:   false,
		ServerCert:    nil,
		ServerKey:     nil,
	}
	cp, _ := collector.InitCollectingProcess(cpInput)

	apInput := intermediate.AggregationInput{
		MessageChan:       cp.GetMsgChan(),
		WorkerNum:         aggregationWorkerNum,
		CorrelateFields:   correlatefields,
		AggregateElements: aggregatedFields,
	}
	ap, _ := intermediate.InitAggregationProcess(apInput)
	go cp.Start()
	waitForCollectorReady(t, cp)
	go func() {
		collectorAddr, _ := net.ResolveTCPAddr("tcp", cp.GetAddress().String())
		conn, err := net.DialTCP("tcp", nil, collectorAddr)
		if err != nil {
			t.Errorf("TCP Collecting Process does not start correctly.")
		}
		defer conn.Close()
		if isIPv6 {
			conn.Write(templatePacketIPv6)
			conn.Write(dataPacket1IPv6)
			conn.Write(dataPacket2IPv6)
		} else {
			conn.Write(templatePacketIPv4)
			conn.Write(dataPacket1IPv4)
			conn.Write(dataPacket2IPv4)
		}
	}()
	go ap.Start()
	if isIPv6 {
		waitForAggregationToFinish(t, ap, flowKey2)
	} else {
		waitForAggregationToFinish(t, ap, flowKey1)
	}
	cp.Stop()
	ap.Stop()

	var record entities.Record
	if isIPv6 {
		assert.NotNil(t, flowKeyRecordMap[flowKey2])
		record = flowKeyRecordMap[flowKey2].Record
	} else {
		assert.NotNil(t, flowKeyRecordMap[flowKey1])
		record = flowKeyRecordMap[flowKey1].Record
	}
	assert.Equal(t, 25, len(record.GetOrderedElementList()))
	for _, element := range record.GetOrderedElementList() {
		switch element.Element.Name {
		case "sourcePodName":
			assert.Equal(t, "pod1", element.Value)
		case "destinationPodName":
			assert.Equal(t, "pod2", element.Value)
		case "flowEndSeconds":
			assert.Equal(t, uint32(1257896000), element.Value)
		case "packetTotalCount":
			assert.Equal(t, uint64(1000), element.Value)
		case "packetDeltaCount":
			assert.Equal(t, uint64(1000), element.Value)
		case "destinationClusterIPv4":
			assert.Equal(t, net.IP{10, 0, 0, 3}, element.Value)
		case "destinationClusterIPv6":
			assert.Equal(t, net.IP{0x20, 0x1, 0x0, 0x0, 0x32, 0x38, 0xbb, 0xbb, 0x0, 0x63, 0x0, 0x0, 0x0, 0x0, 0xaa, 0xaa}, element.Value)
		case "destinationServicePort":
			assert.Equal(t, uint16(4739), element.Value)
		case "reversePacketDeltaCount":
			assert.Equal(t, uint64(350), element.Value)
		case "reversePacketTotalCount":
			assert.Equal(t, uint64(400), element.Value)
		case "packetTotalCountFromSourceNode":
			assert.Equal(t, uint64(800), element.Value)
		case "packetDeltaCountFromSourceNode":
			assert.Equal(t, uint64(500), element.Value)
		case "packetTotalCountFromDestinationNode":
			assert.Equal(t, uint64(1000), element.Value)
		case "packetDeltaCountFromDestinationNode":
			assert.Equal(t, uint64(500), element.Value)
		case "reversePacketTotalCountFromSourceNode":
			assert.Equal(t, uint64(300), element.Value)
		case "reversePacketDeltaCountFromSourceNode":
			assert.Equal(t, uint64(150), element.Value)
		case "reversePacketTotalCountFromDestinationNode":
			assert.Equal(t, uint64(400), element.Value)
		case "reversePacketDeltaCountFromDestinationNode":
			assert.Equal(t, uint64(200), element.Value)
		}
	}

}

func copyFlowKeyRecordMap(key intermediate.FlowKey, aggregationFlowRecord intermediate.AggregationFlowRecord) error {
	flowKeyRecordMap[key] = aggregationFlowRecord
	return nil
}

func waitForCollectorReady(t *testing.T, cp *collector.CollectingProcess) {
	checkConn := func() (bool, error) {
		if strings.Split(cp.GetAddress().String(), ":")[1] == "0" {
			return false, fmt.Errorf("random port is not resolved")
		}
		if _, err := net.Dial(cp.GetAddress().Network(), cp.GetAddress().String()); err != nil {
			return false, err
		}
		return true, nil
	}
	if err := wait.Poll(100*time.Millisecond, 500*time.Millisecond, checkConn); err != nil {
		t.Errorf("Cannot establish connection to %s", cp.GetAddress().String())
	}
}

func waitForAggregationToFinish(t *testing.T, ap *intermediate.AggregationProcess, key intermediate.FlowKey) {
	checkConn := func() (bool, error) {
		ap.ForAllRecordsDo(copyFlowKeyRecordMap)
		if len(flowKeyRecordMap) > 0 {
			ie1, _ := flowKeyRecordMap[key].Record.GetInfoElementWithValue("sourcePodName")
			ie2, _ := flowKeyRecordMap[key].Record.GetInfoElementWithValue("destinationPodName")
			if ie1.Value == "pod1" && ie2.Value == "pod2" {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return false, fmt.Errorf("aggregation process does not process and store data correctly")
		}
	}
	if err := wait.Poll(100*time.Millisecond, 500*time.Millisecond, checkConn); err != nil {
		t.Error(err)
	}
}
