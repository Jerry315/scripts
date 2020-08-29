package bitmap

import (
	"sync"
)

const (
	// initIndex 初始分配的索引值
	initIndex = 1

	// Unlimited 不限制长度
	Unlimited = 0
)

var (
	// NextZeroMaxTries 获取下个为 0 的位尝试次数
	NextZeroMaxTries = 10
)

// Bitmap 位图数据类型, 用于分配和检测索引
type Bitmap struct {
	bits    map[uint32]uint64
	length  uint64
	lastIdx uint64
	mutex   *sync.RWMutex
}

// NewBitmap 创建位图数据
func NewBitmap(length uint64) *Bitmap {
	return &Bitmap{
		bits:   map[uint32]uint64{},
		length: length,
		mutex:  new(sync.RWMutex),
	}
}

// Set 设置指定位为 1
func (bm *Bitmap) Set(idx uint64) {
	if (idx > bm.length || idx == 0) && bm.length != Unlimited {
		return
	}

	word, bit := bm.wordAndBit(idx)
	bm.mutex.Lock()
	if _, ok := bm.bits[word]; !ok {
		bm.bits[word] = 0
	}
	bm.bits[word] |= 1 << bit
	// 设置最后置 1 的索引值, 用于方便查找下一个非 0 位置
	bm.lastIdx = idx
	bm.mutex.Unlock()
}

// Test 检测指定位是否位 1
func (bm *Bitmap) Test(idx uint64) bool {
	if (idx > bm.length || idx == 0) && bm.length != Unlimited {
		return false
	}

	word, bit := bm.wordAndBit(idx)
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()
	if _, ok := bm.bits[word]; !ok {
		return false
	}
	maskedNum := bm.bits[word] & (1 << bit)
	return maskedNum>>bit == 1
}

// TestAndSet 检测指定位是否为 1, 并且设置为 1
func (bm *Bitmap) TestAndSet(idx uint64) bool {
	if (idx > bm.length || idx == 0) && bm.length != Unlimited {
		return false
	}

	isSet := bm.Test(idx)
	bm.Set(idx)
	return isSet
}

// Seek 移动最后位置, 获取下一个 0 值会从 idx+1 开始
func (bm *Bitmap) Seek(idx uint64) {
	if (idx > bm.length || idx == 0) && bm.length != Unlimited {
		return
	}

	bm.mutex.Lock()
	bm.lastIdx = idx
	bm.mutex.Unlock()
}

// Position 返回当前的位索引
func (bm *Bitmap) Position() uint64 {
	bm.mutex.Lock()
	position := bm.lastIdx
	bm.mutex.Unlock()
	return position
}

// Clear 清除指定位, 设置为 0
func (bm *Bitmap) Clear(idx uint64) {
	if (idx > bm.length || idx == 0) && bm.length != Unlimited {
		return
	}

	word, bit := bm.wordAndBit(idx)
	bm.mutex.Lock()
	if _, ok := bm.bits[word]; !ok {
		return
	}
	bm.bits[word] &^= 1 << bit
	bm.mutex.Unlock()
}

// TryAndSetNextZero 尝试并设置下一个
func (bm *Bitmap) TryAndSetNextZero(times int) (uint64, bool) {
	var idx uint64
	var ok bool

	for i := 0; i < times; i++ {
		idx, ok = bm.NextZero()
		if !ok {
			continue
		}
		used := bm.TestAndSet(idx)
		if used {
			continue
		}
		ok = true
		break
	}
	return idx, ok
}

// NextZero 获取下一个 0 位
func (bm *Bitmap) NextZero() (uint64, bool) {
	bm.mutex.Lock()
	nextIdx := bm.lastIdx + 1
	bm.mutex.Unlock()

	// 超出长度, 从初始开始, 注意需要区分长度无限制的情况
	if bm.length != Unlimited && nextIdx > bm.length {
		nextIdx = initIndex
	}
	beginIdx := nextIdx

	tries := 0
	// 循环十次查找值为 0 的位, 如果找不到则返回 0 并通知失败
	for bm.Test(nextIdx) {
		nextIdx++
		// 注意需要区分长度无限制的情况
		if bm.length != Unlimited && nextIdx > bm.length {
			nextIdx = initIndex
		}

		// 回溯到最初增加尝试次数
		if nextIdx == beginIdx {
			tries++
		}
		if tries > NextZeroMaxTries {
			return 0, false
		}
	}
	return nextIdx, true
}

// wordAndBit 获取指定位的字序号和字内位序号
func (bm *Bitmap) wordAndBit(idx uint64) (uint32, uint8) {
	return uint32(idx / 64), uint8(idx % 64)
}
