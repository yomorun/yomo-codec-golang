package varint

import (
	"github.com/yomorun/yomo-codec-golang/internal/utils"
)

// MSB 描述了`1000 0000`, 用于表示后续字节仍然是该变长类型值的一部分
const MSB byte = 0x80

// DropMSB 描述了`0111 1111`, 用于去除标识位使用
const DropMSB = 0x7F

// Varint 定义了一种描述整数的方法, 它是变长类型
//
// 特点：
// 1. 如果当前byte的highest bit是1，则下一个byte也是Varint的一部分。而这个highest bit我们称为MSB
// 2. 如果当前byte的highest bit是0，则该byte是Varint的最后一个部分
// 3. 去掉每一个byte的最高位，按小端序计算
// 4. 使用zigzag算法计算出结果
type Varint struct {
	raw []byte
}

// Decoder 用于解码
type Decoder struct {
	raw    []byte
	logger utils.Logger
}

// NewDecoder return a decoder for parsing `Varint`
func NewDecoder(buf []byte, startPos int) (dec *Decoder, len int) {
	// 因为是可变类型，所以此时还不知道有buf中有多少bytes是Varint使用的
	// 根据Varint类型的特点来寻找Varint的结束位置
	// 并将属于Varint类型的buffer传入Decoder中
	len = 0
	raw := make([]byte, 0)
	for _, b := range buf {
		len++
		raw = append(raw, b)
		if b&MSB != MSB {
			break
		}
	}

	return &Decoder{
		raw:    raw,
		logger: utils.Logger.WithPrefix(utils.DefaultLogger, "varint::Decode"),
	}, len
}

// Decode parse bytes and returns the `uint64` value
func (d *Decoder) Decode() (int64, error) {
	var val uint64
	for i, v := range d.raw {
		val |= (uint64(v & DropMSB)) << (i * 7)
	}
	res := zigzagDecode(val)
	return res, nil
}

/// zigzag from : https://developers.google.com/protocol-buffers/docs/encoding#signed-integers
func zigzagEncode(from int64) uint64 {
	return uint64((from << 1) ^ (from >> 63))
}

/// zigzag from : https://developers.google.com/protocol-buffers/docs/encoding#signed-integers
func zigzagDecode(from uint64) int64 {
	return int64((from >> 1) ^ uint64(-(int64(from & 1))))
}
