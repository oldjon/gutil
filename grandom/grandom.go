package grandom

import (
	"errors"
	"math/rand"
	"reflect"
)

var (
	errRandomPoolEmpty      = errors.New("random pool empty")
	errRandomPoolTypeWrong  = errors.New("random pool need to be slice")
	errRandomWeightZero     = errors.New("random weight is zero")
	errWeightFuncNil        = errors.New("weight func is nil")
	errRandomDstTypeWrong   = errors.New("random dst need to be ptr of slice")
	errRandomPoolAndDstType = errors.New("random pool and dst has different type")
)

// RandFromSlice random out a slice node from the offered slice by weight
func RandFromSlice(pool interface{}, dst interface{}, weight func(i int) uint32) error {
	if weight == nil {
		return errWeightFuncNil
	}
	pv := reflect.ValueOf(pool)
	if pv.Kind() != reflect.Slice {
		return errRandomPoolTypeWrong
	}
	pl := pv.Len()
	if pl == 0 {
		return errRandomPoolEmpty
	}
	dt := reflect.ValueOf(dst)
	if dt.Kind() != reflect.Ptr {
		return errRandomDstTypeWrong
	}

	dte := dt.Elem()
	p0t := pv.Index(0).Type()
	var caseNum int
	if p0t == dt.Type() {
		caseNum = 1
	} else if p0t == dte.Type() {
		caseNum = 2
	} else {
		return errRandomPoolAndDstType
	}

	var totalWeight uint32
	for i := 0; i < pl; i++ {
		totalWeight += weight(i)
	}

	if totalWeight == 0 {
		return errRandomWeightZero
	}

	tmp := uint32(rand.Int63n(int64(totalWeight)) + 1)
	for i := 0; i < pl; i++ {
		if tmp > weight(i) {
			tmp -= weight(i)
			continue
		}
		if caseNum == 1 {
			dte.Set(pv.Index(i).Elem())
		} else {
			dte.Set(pv.Index(i))
		}
		break
	}
	return nil
}

// RandFromSliceDeduplicated random out a sub slice from the offered slice by weight
func RandFromSliceDeduplicated(pool interface{}, needNum uint32, dst interface{}, weight func(i int) uint32) error {
	if weight == nil {
		return errWeightFuncNil
	}
	pv := reflect.ValueOf(pool)
	if pv.Kind() != reflect.Slice {
		return errRandomPoolTypeWrong
	}
	pl := pv.Len()
	if pl == 0 {
		return errRandomPoolEmpty
	}

	dt := reflect.TypeOf(dst)
	if dt.Kind() != reflect.Ptr {
		return errRandomDstTypeWrong
	}

	dv := reflect.ValueOf(dst).Elem()
	if dv.Kind() != reflect.Slice {
		return errRandomDstTypeWrong
	}

	pvt := pv.Type()
	if pvt != dv.Type() {
		return errRandomPoolAndDstType
	}

	var totalWeight uint32
	for i := 0; i < pl; i++ {
		totalWeight += weight(i)
	}

	if totalWeight == 0 {
		return errRandomWeightZero
	}

	tmpDst := reflect.MakeSlice(pvt, int(needNum), int(needNum))
	outed := make([]byte, (pl+7)/8)
	n := needNum
	for ; n > 0; n-- {
		tmp := uint32(rand.Int63n(int64(totalWeight)) + 1)
		for i := 0; i < pl; i++ {
			if outed[i/8]&(1<<(i%8)) > 0 {
				continue
			}
			if tmp > weight(i) {
				tmp -= weight(i)
				continue
			}
			tmpDst.Index(int(needNum - n)).Set(pv.Index(i))
			totalWeight -= weight(i)
			if totalWeight == 0 {
				dv.Set(tmpDst.Slice(0, int(needNum-n)))
				return nil
			}
			outed[i/8] = outed[i/8] | (1 << (i % 8))
			break
		}
	}
	dv.Set(tmpDst.Slice(0, int(needNum-n)))
	return nil
}
