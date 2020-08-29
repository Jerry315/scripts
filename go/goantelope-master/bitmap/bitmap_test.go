package bitmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testSetCases = []struct {
		bitmap          *Bitmap
		length          uint64
		setIdx          uint64
		lastIdxShouldBe uint64
		setOk           bool
	}{
		{NewBitmap(99), 99, 0, 0, false},
		{NewBitmap(99), 99, 88, 88, true},
		{NewBitmap(99), 99, 99, 99, true},
		{NewBitmap(99), 99, 101, 0, false},
	}

	testTestCases = []struct {
		bitmap  *Bitmap
		setIdx  uint64
		testIdx uint64
		testOk  bool
	}{
		{NewBitmap(101), 1, 1, true},
		{NewBitmap(101), 99, 99, true},
		{NewBitmap(101), 101, 101, true},
		{NewBitmap(1021111), 22222, 33333, false},
	}

	testTestAndSetCases = []struct {
		bitmap        *Bitmap
		setIdx        uint64
		testAndSetIdx uint64
		testAndSetOk  bool
	}{
		{NewBitmap(101), 1, 1, true},
		{NewBitmap(101), 99, 99, true},
		{NewBitmap(101), 101, 101, true},
		{NewBitmap(1021111), 22222, 33333, false},
		{NewBitmap(1021111), 1021112, 1021112, false},
	}

	testPositionCases = []struct {
		bitmap   *Bitmap
		seekIdx  uint64
		position uint64
	}{
		{NewBitmap(101001), 1, 1},
		{NewBitmap(101001), 9999, 9999},
		{NewBitmap(101001), 101001, 101001},
		{NewBitmap(101001), 101002, 0},
	}

	testClearCases = []struct {
		bitmap     *Bitmap
		setIdx     uint64
		clearIdx   uint64
		isClearSet bool
	}{
		{NewBitmap(102001), 1, 1, true},
		{NewBitmap(102001), 3020, 3020, true},
		{NewBitmap(102001), 102001, 102001, true},
		{NewBitmap(102001), 1, 102002, false},
	}

	testNextZeroCases = []struct {
		bitmap     *Bitmap
		setIdxs    []uint64
		seekIdx    uint64
		nextZero   uint64
		nextZeroOk bool
	}{
		{NewBitmap(10), []uint64{2, 3, 5}, 1, 4, true},
		{NewBitmap(10), []uint64{}, 5, 6, true},
		{NewBitmap(10), []uint64{2, 3, 4}, 10, 1, true},
		{NewBitmap(5), []uint64{1, 2, 3, 4, 5}, 1, 0, false},
	}

	testTryAndSetNextZeroCases = []struct {
		bitmap      *Bitmap
		setIdxs     []uint64
		seekIdx     uint64
		nextZero    uint64
		nextZeroOk  bool
		nnextZero   uint64
		nnextZeroOK bool
	}{
		{NewBitmap(10), []uint64{2, 3, 5}, 1, 4, true, 6, true},
		{NewBitmap(10), []uint64{}, 5, 6, true, 7, true},
		{NewBitmap(10), []uint64{2, 3, 4}, 10, 1, true, 5, true},
		{NewBitmap(5), []uint64{1, 2, 3, 4, 5}, 1, 0, false, 0, false},
		{NewBitmap(Unlimited), []uint64{}, 0, 1, true, 2, true},
	}
)

func TestNewBitmap(t *testing.T) {
	assert := assert.New(t)

	bitmap := NewBitmap(101)
	assert.NotNil(bitmap)
	assert.Equal(bitmap.length, uint64(101))
	assert.NotNil(bitmap.mutex)
	assert.NotNil(bitmap.bits)
	assert.Equal(bitmap.lastIdx, uint64(0))
}

func TestSet(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range testSetCases {
		assert.Equal(tc.bitmap.lastIdx, uint64(0))
		assert.Equal(tc.bitmap.length, tc.length)
		assert.Equal(len(tc.bitmap.bits), 0)

		tc.bitmap.Set(tc.setIdx)
		assert.Equal(tc.lastIdxShouldBe, tc.bitmap.lastIdx)
		word, bit := wordAndBit(tc.setIdx)
		if tc.setOk {
			assert.Equal(len(tc.bitmap.bits), 1)
		}

		wordNum, ok := tc.bitmap.bits[word]
		assert.Equal(ok, tc.setOk)
		assert.Equal(isNumBitOne(wordNum, bit), tc.setOk)
	}
}

func TestTest(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range testTestCases {
		tc.bitmap.Set(tc.setIdx)
		assert.Equal(tc.bitmap.Test(tc.testIdx), tc.testOk)
	}
}

func TestTestAndSet(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range testTestAndSetCases {
		tc.bitmap.Set(tc.setIdx)
		assert.Equal(tc.bitmap.TestAndSet(tc.testAndSetIdx), tc.testAndSetOk)
		if tc.testAndSetOk {
			assert.Equal(tc.bitmap.Test(tc.testAndSetIdx), true)
		}
	}
}

func TestSeek(t *testing.T) {
	assert := assert.New(t)

	length := uint64(101)
	bitmap := NewBitmap(length)
	for idx := uint64(1); idx <= length; idx++ {
		bitmap.Seek(idx)
		assert.Equal(bitmap.lastIdx, idx)
	}
}

func TestPosition(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range testPositionCases {
		tc.bitmap.Seek(tc.seekIdx)
		assert.Equal(tc.bitmap.Position(), tc.position)
	}
}

func TestClear(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range testClearCases {
		tc.bitmap.Set(tc.setIdx)
		assert.Equal(tc.bitmap.Test(tc.setIdx), true)
		tc.bitmap.Clear(tc.clearIdx)
		assert.Equal(tc.bitmap.Test(tc.clearIdx), false)
		if tc.isClearSet {
			assert.Equal(tc.bitmap.Test(tc.setIdx), false)
		} else {
			assert.Equal(tc.bitmap.Test(tc.setIdx), true)
		}
	}
}

func TestNextZero(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range testNextZeroCases {
		for _, idx := range tc.setIdxs {
			tc.bitmap.Set(idx)
		}
		tc.bitmap.Seek(tc.seekIdx)
		nextZero, ok := tc.bitmap.NextZero()

		assert.Equal(ok, tc.nextZeroOk)
		if tc.nextZeroOk {
			assert.Equal(nextZero, tc.nextZero)
		}
	}
}

func TestUnlimitedBitmap(t *testing.T) {
	assert := assert.New(t)

	bitmap := NewBitmap(Unlimited)
	assert.Equal(bitmap.length, uint64(0))

	idxs := []uint64{1, 100001, 2000000002, 3000000000000003, 40000000000000004}
	for _, idx := range idxs {
		bitmap.Set(idx)
		assert.Equal(bitmap.Test(idx), true)
	}
}

func TestTryAndSetNextZero(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range testTryAndSetNextZeroCases {
		for _, idx := range tc.setIdxs {
			tc.bitmap.Set(idx)
		}
		tc.bitmap.Seek(tc.seekIdx)
		idx, ok := tc.bitmap.TryAndSetNextZero(3)

		assert.Equal(ok, tc.nextZeroOk)
		if tc.nextZeroOk {
			assert.Equal(tc.bitmap.Test(idx), true)
			assert.Equal(tc.nextZero, idx)
		}

		nnextIdx, ok := tc.bitmap.TryAndSetNextZero(3)
		assert.Equal(ok, tc.nnextZeroOK)
		if tc.nnextZeroOK {
			assert.Equal(tc.bitmap.Test(nnextIdx), true)
			assert.Equal(tc.nnextZero, nnextIdx)
		}
	}
}

func wordAndBit(idx uint64) (uint32, uint8) {
	return uint32(idx / 64), uint8(idx % 64)
}

// isNumBitOne 检查数字指定位是否为 1
func isNumBitOne(num uint64, bit uint8) bool {
	maskedNum := num & (1 << bit)
	return maskedNum>>bit == 1
}
