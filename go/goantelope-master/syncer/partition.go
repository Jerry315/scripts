package syncer

import (
	"sync"

	rbtree "github.com/emirpasic/gods/trees/redblacktree"
)

// Partition CID 段数据
type Partition struct {
	Begin uint32 `json:"start"`
	End   uint32 `json:"end"`
}

// Partitions CID 段集合
type Partitions struct {
	partitionMutex sync.RWMutex
	treeMutex      sync.RWMutex
	tree           *rbtree.Tree
}

// NewPartitions 新建 Partitions 变量
func NewPartitions() *Partitions {
	return &Partitions{
		partitionMutex: sync.RWMutex{},
		treeMutex:      sync.RWMutex{},
		tree:           rbtree.NewWith(PartitionComparator),
	}
}

// Find 查找 CID 所属 CID 段
func (p *Partitions) Find(cid uint32) (string, bool) {
	partition := Partition{
		Begin: cid,
		End:   cid,
	}

	p.treeMutex.RLock()
	value, found := p.tree.Get(partition)
	p.treeMutex.RUnlock()

	if !found {
		return "", false
	}
	valueStr, ok := value.(string)
	if !ok {
		return "", false
	}
	return valueStr, true
}

// Update 更新 CID 段数据
func (p *Partitions) Update(appID string, partitions []Partition) {
	p.treeMutex.Lock()
	defer p.treeMutex.Unlock()

	for _, partition := range partitions {
		p.tree.Put(partition, appID)
	}
}

// Remove 移除 CID 段数据
func (p *Partitions) Remove(begin, end uint32) {
	partition := Partition{Begin: begin, End: end}
	p.tree.Remove(partition)
}

// Build 构建 CID 段数据的查找树, 会替换旧的树
func (p *Partitions) Build() {
	// 留作兼容函数
	return
}

// Clear 清除当前的 CID 段数据
func (p *Partitions) Clear() {
	p.treeMutex.Lock()
	defer p.treeMutex.Unlock()
	p.tree = rbtree.NewWith(PartitionComparator)
}

// PartitionComparator 两个 CID 段对大小对比函数
func PartitionComparator(a, b interface{}) int {
	p1 := a.(Partition)
	p2 := b.(Partition)

	switch {
	case (p1.Begin > p2.Begin) && (p1.End > p2.End):
		return 1
	case (p1.Begin < p2.Begin) && (p1.End < p2.End):
		return -1
	default:
		return 0
	}
}
