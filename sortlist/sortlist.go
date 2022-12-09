package sortlist

type SortNode struct {
	Key   uint64
	Value interface{}
}

type SortList struct {
	SList []SortNode
}

func (sl *SortList) Insert(key uint64, value interface{}) {
	var (
		lSize = len(sl.SList)
		low   = 0
		high  = lSize - 1
		mid   int
		node  *SortNode
	)
	if lSize == 0 || sl.SList[high].Key < key {
		sl.SList = append(sl.SList, SortNode{
			Key:   key,
			Value: value,
		})
		return
	}
	for low <= high {
		mid = (low + high) >> 1
		node = &sl.SList[mid]
		if node.Key == key {
			node.Value = value
			return
		}
		if node.Key > key {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	index := mid
	if node.Key < key {
		index = mid + 1
	}
	sl.SList = append(sl.SList, SortNode{})
	copy(sl.SList[index+1:], sl.SList[index:])
	sl.SList[index] = SortNode{
		Key:   key,
		Value: value,
	}
	// sl.SList = append(sl.SList, SortNode{
	// 	Key:   key,
	// 	Value: value,
	// })
	// for i := lSize; i > index; i-- {
	// 	sl.SList[i], sl.SList[i-1] = sl.SList[i-1], sl.SList[i]
	// }
}

func (sl *SortList) Find(key uint64) interface{} {
	lSize := len(sl.SList)
	if lSize == 0 {
		return nil
	}
	var (
		low  = 0
		high = lSize - 1
		mid  int
		node *SortNode
	)
	for low <= high {
		mid = (low + high) >> 1
		node = &sl.SList[mid]
		if node.Key == key {
			return node.Value
		}
		if node.Key > key {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return nil
}

func (sl *SortList) Erase(key uint64) int {
	lSize := len(sl.SList)
	if lSize == 0 {
		return -1
	}
	var (
		low  = 0
		high = lSize - 1
		mid  int
		node *SortNode
	)
	for low <= high {
		mid = (low + high) >> 1
		node = &sl.SList[mid]
		if node.Key == key {
			sl.SList = append(sl.SList[:mid], sl.SList[mid+1:]...)
			return mid
		}
		if node.Key > key {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return -1
}

func (sl *SortList) Clear() {
	sl.SList = sl.SList[:0]
}

func (sl *SortList) Size() int {
	return len(sl.SList)
}

func (sl *SortList) List() []SortNode {
	return sl.SList
}
