package syncer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	partitionData = []struct {
		appID      string
		partitions []Partition
	}{
		{
			"test",
			[]Partition{
				Partition{Begin: 100041, End: 100051},
				Partition{Begin: 100061, End: 100071},
				Partition{Begin: 100081, End: 100091},
				Partition{Begin: 1000, End: 2000},
			},
		},
		{
			"xybk",
			[]Partition{
				Partition{Begin: 3000, End: 4000},
				Partition{Begin: 200000, End: 300000},
			},
		},
	}

	cidData = []struct {
		cid   uint32
		appID string
		ok    bool
	}{
		{100041, "test", true},
		{100042, "test", true},
		{100091, "test", true},
		{2001, "", false},
		{3500, "xybk", true},
		{250000, "xybk", true},
		{19, "", false},
	}
)

func TestUpdateAndClear(t *testing.T) {
	assert := assert.New(t)

	partitions := NewPartitions()
	for _, data := range partitionData {
		oldCnt := testUtilPartitionCnt(partitions)
		incCnt := len(data.partitions)

		partitions.Update(data.appID, data.partitions)
		curCnt := testUtilPartitionCnt(partitions)
		assert.Equal(incCnt, curCnt-oldCnt)
	}

	partitions.Clear()
	assert.Equal(0, testUtilPartitionCnt(partitions))
}

func TestFind(t *testing.T) {
	assert := assert.New(t)

	partitions := testUtilInitTestPartitions()
	partitions.Build()
	for _, data := range cidData {
		appID, found := partitions.Find(data.cid)
		assert.Equal(data.ok, found)
		assert.Equal(data.appID, appID)
	}

	// test fail data
	partitions.tree.Put(Partition{Begin: 1, End: 10}, 234)
	value, found := partitions.Find(5)
	assert.Equal(false, found)
	assert.Equal("", value)
}

func testUtilPartitionCnt(p *Partitions) int {
	return p.tree.Size()
}

func testUtilInitTestPartitions() *Partitions {
	partitions := NewPartitions()
	for _, data := range partitionData {
		partitions.Update(data.appID, data.partitions)
	}
	return partitions
}
