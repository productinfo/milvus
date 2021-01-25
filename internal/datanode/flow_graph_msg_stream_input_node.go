package datanode

import (
	"context"

	"github.com/zilliztech/milvus-distributed/internal/msgstream"
	"github.com/zilliztech/milvus-distributed/internal/msgstream/pulsarms"
	"github.com/zilliztech/milvus-distributed/internal/msgstream/util"
	"github.com/zilliztech/milvus-distributed/internal/util/flowgraph"
)

func newDmInputNode(ctx context.Context) *flowgraph.InputNode {
	receiveBufSize := Params.InsertReceiveBufSize
	pulsarBufSize := Params.InsertPulsarBufSize

	msgStreamURL := Params.PulsarAddress

	consumeChannels := Params.InsertChannelNames
	consumeSubName := Params.MsgChannelSubName

	insertStream := pulsarms.NewPulsarTtMsgStream(ctx, receiveBufSize)

	insertStream.SetPulsarClient(msgStreamURL)
	unmarshalDispatcher := util.NewUnmarshalDispatcher()

	insertStream.CreatePulsarConsumers(consumeChannels, consumeSubName, unmarshalDispatcher, pulsarBufSize)

	var stream msgstream.MsgStream = insertStream

	maxQueueLength := Params.FlowGraphMaxQueueLength
	maxParallelism := Params.FlowGraphMaxParallelism

	node := flowgraph.NewInputNode(&stream, "dmInputNode", maxQueueLength, maxParallelism)
	return node
}

func newDDInputNode(ctx context.Context) *flowgraph.InputNode {

	consumeChannels := Params.DDChannelNames
	consumeSubName := Params.MsgChannelSubName

	ddStream := pulsarms.NewPulsarTtMsgStream(ctx, 1024)
	ddStream.SetPulsarClient(Params.PulsarAddress)
	unmarshalDispatcher := util.NewUnmarshalDispatcher()
	ddStream.CreatePulsarConsumers(consumeChannels, consumeSubName, unmarshalDispatcher, 1024)

	var stream msgstream.MsgStream = ddStream

	maxQueueLength := Params.FlowGraphMaxQueueLength
	maxParallelism := Params.FlowGraphMaxParallelism

	node := flowgraph.NewInputNode(&stream, "ddInputNode", maxQueueLength, maxParallelism)
	return node
}