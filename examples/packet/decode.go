package main

import (
	"fmt"

	"github.com/yomorun/y3-codec-golang/examples"

	"github.com/yomorun/y3-codec-golang/internal/utils"

	y3 "github.com/yomorun/y3-codec-golang"
)

func main() {
	fmt.Println("hello YoMo YomoCodec golang implementation: Y3")
	encodePacket()
	encodeArrayPacket()

	parseNodePacket()
	parseComplexNodePacket()

	parseInt32PrimitivePacket()
	parseUInt32PrimitivePacket()
	parseArrayPacket()
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
	bars  []*bar
	kinds []int32
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
	var bars = y3.NewNodeSlicePacketEncoder(0x02)

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
	var kinds = y3.NewNodeSlicePacketEncoder(0x04)

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
	fmt.Printf("club=%#v\n", buf)

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
	examples.PrintNodePacket(node)
}

func parseInt32PrimitivePacket() {
	fmt.Println(">> Parsing [0x0A, 0x01, 0x7F] EQUALS key-value = 0x0A: 127")
	buf := []byte{0x0A, 0x01, 0x7F}
	res, _, _, err := y3.DecodePrimitivePacket(buf)
	v1, err := res.ToInt32()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tag Key=[%#X], Value=%v\n", res.SeqID(), v1)
}

func parseUInt32PrimitivePacket() {
	fmt.Println(">> Parsing [0x0A, 0x02, 0x80, 0x7F], which like Key-Value format = 0x0A: 127")
	buf := []byte{0x0A, 0x02, 0x80, 0x7F}
	res, _, _, err := y3.DecodePrimitivePacket(buf)
	v1, err := res.ToUInt32()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tag Key=[%#X], Value=%v\n", res.SeqID(), v1)
}

func printArrayNode(node *y3.NodePacket) {
	examples.PrintNodePacket(node)
}

func parseArrayPacket() {
	utils.DefaultLogger.SetLogLevel(utils.LogLevelDebug)
	/*
		0xc1:[ { 0x03:0x61 }, { 0x04:0x62 } ]
	*/
	fmt.Println(">> Parsing [0xc1, 0x06, 0x03, 0x01, 0x61, 0x04, 0x01, 0x62] EQUALS 0xc1:[{0x03:0x61},{0x04:0x62}]")
	buf := []byte{0xc1, 0x06, 0x03, 0x01, 0x61, 0x04, 0x01, 0x62}
	//res, _, _, _ := y3.DecodePrimitivePacket(buf)
	res, _, _ := y3.DecodeNodePacket(buf)
	printArrayNode(res)
	println()

	/*
		0xc1:[ 0x00:0x61, 0x00:0x62 ]
	*/
	fmt.Println(">> Parsing [0xc1, 0x06, 0x00, 0x01, 0x61, 0x00, 0x01, 0x62] EQUALS 0xc1:[0x02,0x04]")
	buf = []byte{0xc1, 0x06, 0x00, 0x01, 0x02, 0x00, 0x01, 0x04}
	//res, _, _, _ = y3.DecodePrimitivePacket(buf)
	res, _, _ = y3.DecodeNodePacket(buf)
	//printArray(res)
	examples.PrintArrayPacket(res)
}

func parseNestedArrayPrimitivePacket() {
	utils.DefaultLogger.SetLogLevel(utils.LogLevelDebug)
	/*
		0x41:[
			0x42: [
				0x00: 0x61 0x31,
				0x00: 0x62 0x32
			],
			0x05: 0x63
		]
	*/
	fmt.Println(">> Parsing [0xc1, 0x0d, 0xc2, 0x08, 0x00, 0x02, 0x61, 0x31, 0x00, 0x02, 0x62, 0x32, 0x05, 0x01, 0x63] EQUALS 0x41:[0x42:[0x6131,0x6232],0x05:0x63]")
	buf := []byte{0xc1, 0x0d, 0xc2, 0x08, 0x00, 0x02, 0x61, 0x31, 0x00, 0x02, 0x62, 0x32, 0x05, 0x01, 0x63}
	//fmt.Println("#5", "len(buf):", len(buf))
	//res, _, _, _ := y3.DecodePrimitivePacket(buf)
	res, _, _ := y3.DecodeNodePacket(buf)
	printNodePacket(res)
}
