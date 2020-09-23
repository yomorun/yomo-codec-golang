package main

import (
	"fmt"

	"github.com/yomorun/yomo-codec-golang/internal/utils"

	y3 "github.com/yomorun/yomo-codec-golang"
)

func main() {
	fmt.Println("hello YoMo Codec golang implementation: Y3")
	encodePacket()
	encodeArrayPacket()

	parseNodePacket()
	parseComplexNodePacket()

	parseStringPrimitivePacket()
	parseArrayPrimitivePacket()
	parseNestedArrayPrimitivePacket()
}

type bar struct {
	Name string
}

type foo struct {
	ID int
	*bar
}

type club struct {
	bars []*bar
}

func encodePacket() {
	// We will encode JSON-like object `obj`:
	// 0x81: {
	//   0x02: -1,
	//   0x83 : {
	//     0x04: "C",
	//   },
	// }
	// to
	// [0x81, 0x08, 0x02, 0x01, 0x7F, 0x83, 0x03, 0x04, 0x01, 0x43]
	var obj = &foo{ID: -1, bar: &bar{Name: "C"}}

	// 0x81 - node
	var foo = y3.NewNodePacketEncoder(0x01)

	// 0x02 - ID=-1
	var yp1 = y3.NewPrimitivePacketEncoder(0x02)
	yp1.SetInt32Value(-1)
	foo.AddPrimitivePacket(yp1)

	// 0x83 - &bar{}
	var bar = y3.NewNodePacketEncoder(0x03)

	// 0x04 - Name="C"
	var yp2 = y3.NewPrimitivePacketEncoder(0x04)
	yp2.SetStringValue("C")
	bar.AddPrimitivePacket(yp2)

	foo.AddNodePacket(bar)

	fmt.Printf("obj=%#v\n", obj)
	fmt.Printf("res=%#v\n", foo.Encode())
}

func encodeArrayPacket() {
	/*
		{
			"club": {
				"bars": [{
					"Name": "a1"
				}, {
					"Name": "a2"
				}],
				"kinds": [10, 11]
			}
		}
		-->
		0x81:
			0xc2:
				0x80: 0x03: 0x61 0x31
				0x80: 0x03: 0x61 0x32
			0xc4:
				0x00: 0x80, 0x7F
				0x00: 0x81, 0x7F
		-->
		0x81, [0x0e], 0xc2, [0x0c], 0x80, [0x04], 0x03, [0x02], 0x61 0x31, 0x80, [0x04], 0x03, [0x02], 0x61 0x32
		0x81, [0x18], 0xc2, [0x0c], 0x80, [0x04], 0x03, [0x02], 0x61 0x31, 0x80, [0x04], 0x03, [0x02], 0x61 0x32, 0xc4, [0x08], 0x00, [0x02], 0x80, 0x7F, 0x00, [0x02], 0x81, 0x7F
	*/
	var obj = &club{bars: []*bar{{Name: "a1"}, {Name: "a2"}}}

	// 0x81 - node
	var club = y3.NewNodePacketEncoder(0x01)

	// 0xc2 - []*bar
	var bars = y3.NewNodeArrayPacketEncoder(0x02)

	// 0x03 - Name="a1"
	var bar1 = y3.NewNodePacketEncoder(0x00)
	var a1 = y3.NewPrimitivePacketEncoder(0x03)
	a1.SetStringValue("a1")
	bar1.AddPrimitivePacket(a1)
	bars.AddNodePacket(bar1)

	// 0x03 - Name="a2"
	var bar2 = y3.NewNodePacketEncoder(0x00)
	var a2 = y3.NewPrimitivePacketEncoder(0x03)
	a2.SetStringValue("a2")
	bar2.AddPrimitivePacket(a2)
	bars.AddNodePacket(bar2)

	// 0x44 - kinds
	var kinds = y3.NewNodeArrayPacketEncoder(0x04)

	// 0x44 - item1
	var item1 = y3.NewPrimitivePacketEncoder(0x00)
	item1.SetInt32Value(127)
	kinds.AddPrimitivePacket(item1)

	// 0x44 - item2
	var item2 = y3.NewPrimitivePacketEncoder(0x00)
	item2.SetInt32Value(255)
	kinds.AddPrimitivePacket(item2)

	club.AddNodePacket(bars)
	club.AddNodePacket(kinds)

	buf := club.Encode()
	fmt.Printf("obj=%#v\n", obj)
	fmt.Printf("bars=%#v\n", buf)

	res, _, _ := y3.DecodeNodePacket(buf)
	printNodePacket(res)
}

func parseNodePacket() {
	fmt.Println(">> Parsing [0x84, 0x06, 0x0A, 0x01, 0x7F, 0x0B, 0x01, 0x43] EQUALS JSON= 0x84: { 0x0A: -1, 0x0B: 'C' }")
	buf := []byte{0x84, 0x06, 0x0A, 0x01, 0x7F, 0x0B, 0x01, 0x43}
	res, _, err := y3.DecodeNodePacket(buf)
	v1 := res.PrimitivePackets[0]

	p1, err := v1.ToInt32()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Tag Key=[%#X.%#X], Value=%v\n", res.SeqID(), v1.SeqID(), p1)

	v2 := res.PrimitivePackets[1]

	p2, err := v2.ToUTF8String()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tag Key=[%#X.%#X], Value=%v\n", res.SeqID(), v2.SeqID(), p2)
}

func parseComplexNodePacket() {
	/**
	{
		"list": [
			{
				"ID":1,
				"bar":{"Name":"b1"}
			},
			{
				"ID":2,
				"bar":{"Name":"b2"}
			}
		],
		"tags": [1, 2]
	}
	-->
	0x81:
		0xC2:
			0x80:
				0x03: 0x01
				0x84: 0x05: 0x62, 0x31
			0x80:
				0x06: 0x02
				0x87: 0x08: 0x62, 0x32
		0x49:
			0x00: 0x01
			0x00: 0x02

	-->
	0x81, [0x20], 0xC2, [0x16],
						0x80, [0x09],
									0x03, [0x01], 0x01,
									0x84, [0x04], 0x05, [0x02], 0x62, 0x31,
						0x80, [0x09],
									0x06, [0x01], 0x02,
									0x87, [0x04], 0x08, [0x02], 0x62, 0x32
	              0x49, [0x06],
						0x00, [0x01], 0x01
						0x00, [0x01], 0x02
	-->
	0x81, 0x18, 0xC2, 0x16, 0x80, 0x09, 0x03, 0x01, 0x01, 0x84, 0x04, 0x05, 0x02, 0x62, 0x31, 0x80, 0x09, 0x06, 0x01, 0x02, 0x87, 0x04, 0x08, 0x02, 0x62, 0x32
	0x81, 0x20, 0xC2, 0x16, 0x80, 0x09, 0x03, 0x01, 0x01, 0x84, 0x04, 0x05, 0x02, 0x62, 0x31, 0x80, 0x09, 0x06, 0x01, 0x02, 0x87, 0x04, 0x08, 0x02, 0x62, 0x32, 0x49, 0x06, 0x00, 0x01, 0x01, 0x00, 0x01, 0x02
	*/
	utils.DefaultLogger.SetLogLevel(utils.LogLevelDebug)
	//fmt.Println(">> Parsing [0x81, 0x18, 0xC2, 0x16, 0x80, 0x09, 0x03, 0x01, 0x01, 0x84, 0x04, 0x05, 0x02, 0x62, 0x31, 0x80, 0x09, 0x06, 0x01, 0x02, 0x87, 0x04, 0x08, 0x02, 0x62, 0x32] EQUALS")
	//buf := []byte{0x81, 0x18, 0xC2, 0x16, 0x80, 0x09, 0x03, 0x01, 0x01, 0x84, 0x04, 0x05, 0x02, 0x62, 0x31, 0x80, 0x09, 0x06, 0x01, 0x02, 0x87, 0x04, 0x08, 0x02, 0x62, 0x32}
	fmt.Println(">> Parsing [0x81, 0x20, 0xC2, 0x16, 0x80, 0x09, 0x03, 0x01, 0x01, 0x84, 0x04, 0x05, 0x02, 0x62, 0x31, 0x80, 0x09, 0x06, 0x01, 0x02, 0x87, 0x04, 0x08, 0x02, 0x62, 0x32, 0x49, 0x06, 0x00, 0x01, 0x01, 0x00, 0x01, 0x02] EQUALS")
	buf := []byte{0x81, 0x20, 0xC2, 0x16, 0x80, 0x09, 0x03, 0x01, 0x01, 0x84, 0x04, 0x05, 0x02, 0x62, 0x31, 0x80, 0x09, 0x06, 0x01, 0x02, 0x87, 0x04, 0x08, 0x02, 0x62, 0x32, 0x49, 0x06, 0x00, 0x01, 0x01, 0x00, 0x01, 0x02}
	res, _, _ := y3.DecodeNodePacket(buf)

	printNodePacket(res)
}

func printNodePacket(node *y3.NodePacket) {
	if len(node.NodePackets) > 0 {
		for _, n := range node.NodePackets {
			printNodePacket(&n)
		}
	}
	if len(node.PrimitivePackets) > 0 {
		for _, p := range node.PrimitivePackets {
			if p.HasPacketArray() {
				printPacketArray(&p)
				continue
			}
			fmt.Printf("#35 %#X=%v\n", p.SeqID(), valueOf(&p))
		}
	}
}

func parseStringPrimitivePacket() {
	fmt.Println(">> Parsing [0x0A, 0x01, 0x7F] EQUALS key-value = 0x0A: 127")
	buf := []byte{0x0A, 0x01, 0x7F}
	//res, _, err := y3.DecodePrimitivePacket(buf)
	res, _, _, err := y3.DecodePrimitivePacket(buf)
	v1, err := res.ToInt32()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tag Key=[%#X], Value=%v\n", res.SeqID(), v1)
}

func parseArrayPrimitivePacket() {
	utils.DefaultLogger.SetLogLevel(utils.LogLevelDebug)
	/*
		0x41:[ { 0x03:0x61 }, { 0x04:0x62 } ]
	*/
	fmt.Println(">> Parsing [0x41, 0x06, 0x03, 0x01, 0x61, 0x04, 0x01, 0x62] EQUALS 0x41:[{0x03:0x61},{0x04:0x62}]")
	buf := []byte{0x41, 0x06, 0x03, 0x01, 0x61, 0x04, 0x01, 0x62}
	res, _, _, _ := y3.DecodePrimitivePacket(buf)
	printPacketArray(res)

	/*
		0x41:[ 0x00:0x61, 0x00:0x62 ]
	*/
	fmt.Println(">> Parsing [0x41, 0x06, 0x00, 0x01, 0x61, 0x00, 0x01, 0x62] EQUALS 0x41:[{0x00:0x02},{0x00:0x04}]")
	buf = []byte{0x41, 0x06, 0x00, 0x01, 0x02, 0x00, 0x01, 0x04}
	res, _, _, _ = y3.DecodePrimitivePacket(buf)
	arr, _ := res.ToPacketArray()
	for _, item := range arr {
		i, _ := item.ToInt32()
		fmt.Println("#30", "Item:", fmt.Sprintf("value=%v", i))
	}
}

func parseNestedArrayPrimitivePacket() {
	utils.DefaultLogger.SetLogLevel(utils.LogLevelDebug)
	/*
		0x41:[
			0x42: [
				0x03: 0x61 0x31,
				0x04: 0x62 0x32
			],
			0x05: 0x63
		]
	*/
	fmt.Println(">> Parsing [0x41, 0x0d, 0x42, 0x08, 0x03, 0x02, 0x61, 0x31, 0x04, 0x02, 0x62, 0x32, 0x05, 0x01, 0x63] EQUALS 0x41:[0x42:[0x03:0x610x31,0x04:0x620x32],0x05:0x63]")
	buf := []byte{0x41, 0x0d, 0x42, 0x08, 0x03, 0x02, 0x61, 0x31, 0x04, 0x02, 0x62, 0x32, 0x05, 0x01, 0x63}
	//fmt.Println("#5", "len(buf):", len(buf))
	res, _, _, _ := y3.DecodePrimitivePacket(buf)
	printPacketArray(res)
}

func printPacketArray(packet *y3.PrimitivePacket) {
	arr, _ := packet.ToPacketArray()
	//fmt.Println("#20", "len(arr):", len(arr))
	for _, item := range arr {
		if item.HasPacketArray() {
			printPacketArray(item)
		} else {
			fmt.Println("#20", "Item:", fmt.Sprintf("key=%v value=%v", item.SeqID(), valueOf(item)))
		}
	}
}

func valueOf(packet *y3.PrimitivePacket) interface{} {
	num, err := packet.ToInt32()
	if err == nil && num > 0 {
		return num
	}
	str, _ := packet.ToUTF8String()
	return str
}

func int32Of(packet *y3.PrimitivePacket) int32 {
	n, _ := packet.ToInt32()
	return n
}

func stringOf(packet *y3.PrimitivePacket) string {
	str, _ := packet.ToUTF8String()
	return str
}
