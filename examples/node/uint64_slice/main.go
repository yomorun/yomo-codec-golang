package main

import (
	"fmt"

	"github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/y3-codec-golang/internal/utils"
)

// Example of encoding and decoding uint64 slice type by using NodePacket.
func main() {
	// encode
	data := []uint64{123, 456}
	var node = y3.NewNodeSlicePacketEncoder(0x10)
	if out, ok := utils.ToUInt64Slice(data); ok {
		for _, v := range out {
			var item = y3.NewPrimitivePacketEncoder(0x00)
			item.SetUInt64Value(v.(uint64))
			node.AddPrimitivePacket(item)
		}
	}
	buf := node.Encode()
	// decode
	packet, _, _ := y3.DecodeNodePacket(buf)
	result := make([]uint64, 0)
	for _, p := range packet.PrimitivePackets {
		v, _ := p.ToUInt64()
		result = append(result, v)
	}
	fmt.Printf("result=%v", result)
}
