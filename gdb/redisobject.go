package gdb

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

func (rc *redisClient) SetKOMapping(mapping map[string]string) {
	rc.koMapping = make(map[string]string)
	for k, v := range mapping {
		if k == "" {
			continue
		}
		k = strings.Split(k, ":")[0]
		strs := strings.Split(v, ".")
		rc.koMapping[k] = strs[len(strs)-1]
	}
}

func (rc *redisClient) CheckKeyObjMatch(key string, obj any) { // panic if not match
	key = strings.Split(key, ":")[0]
	t, ok := rc.koMapping[key]
	if !ok {
		return
	}
	ot := reflect.TypeOf(obj).String()
	strs := strings.Split(ot, ".")
	if t != strs[len(strs)-1] {
		panic("db key and db object type is not match") // if type is not match, means there is a bug in code
	}
}

func (rc *redisClient) IsErrNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

func (rc *redisClient) GetObject(ctx context.Context, key string, obj any) error {
	rc.CheckKeyObjMatch(key, obj)
	v, err := rc.Get(ctx, key)
	if err != nil {
		return err
	}
	return rc.objMarshaller.Unmarshal([]byte(v), obj)
}

func (rc *redisClient) SetObject(ctx context.Context, key string, obj any) error {
	rc.CheckKeyObjMatch(key, obj)
	bys, err := rc.objMarshaller.Marshal(obj) //测试是否可以传nil指针
	if err != nil {
		return err
	}
	return rc.Set(ctx, key, bys)
}

func (rc *redisClient) SetObjectEX(ctx context.Context, key string, obj any, expiration time.Duration) error {
	rc.CheckKeyObjMatch(key, obj)
	bys, err := rc.objMarshaller.Marshal(obj)
	if err != nil {
		return err
	}
	return rc.SetEX(ctx, key, bys, expiration)
}

func (rc *redisClient) GetObjects(ctx context.Context, keys []string, objs any) error {
	if len(keys) == 0 {
		panic(PanicKeyIsMissing)
	}
	objsValue := reflect.ValueOf(objs)
	if objsValue.Kind() != reflect.Slice {
		panic(PanicValueDstNeedBeSlice)
	}
	if len(keys) != objsValue.Len() {
		panic(PanicKeyValueCountUnmatched)
	}

	var elemIsInterface bool
	for i := 0; i < len(keys); i++ {
		rc.CheckKeyObjMatch(keys[i], objsValue.Index(i).Interface())
		if objsValue.Index(i).Kind() != reflect.Ptr && // input is slice of known struct type
			objsValue.Index(i).Elem().Kind() != reflect.Ptr { // input is slice of interface
			panic(PanicValueDstNeedBePointer)
		}
	}
	if objsValue.Index(0).Kind() != reflect.Ptr {
		elemIsInterface = true
	}

	cmds, err := rc.BatchGet(ctx, keys)
	if err != nil {
		return err
	}
	for i, v := range cmds {
		objv := objsValue.Index(i)
		if rc.IsErrNil(v.Err()) {
			if !elemIsInterface {
				objv.Set(reflect.Zero(objv.Type()))
			} else {
				objv.Set(reflect.Zero(objv.Elem().Type()))
			}
			continue
		} else if v.Err() != nil {
			return err
		}
		if !elemIsInterface {
			if objv.IsNil() {
				objv.Set(reflect.New(objv.Type().Elem()))
			}
		} else {
			if objv.Elem().IsNil() {
				objv.Set(reflect.New(objv.Elem().Type().Elem()))
			}
		}

		err = rc.objMarshaller.Unmarshal([]byte(v.Val()), objv.Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (rc *redisClient) SetObjects(ctx context.Context, keys []string, objs any) error {
	return rc.SetObjectsEX(ctx, keys, objs, 0)
}

func (rc *redisClient) SetObjectsEX(ctx context.Context, keys []string, objs any, expiration time.Duration) error {
	if len(keys) == 0 {
		panic(PanicKeyIsMissing)
	}
	objsValue := reflect.ValueOf(objs)
	if objsValue.Kind() != reflect.Slice {
		panic(PanicValueNeedBeSlice)
	}
	if len(keys) != objsValue.Len() {
		panic(PanicKeyValueCountUnmatched)
	}
	var err error
	datas := make([]any, objsValue.Len())
	for i := 0; i < objsValue.Len(); i++ {
		objv := objsValue.Index(i).Interface()
		rc.CheckKeyObjMatch(keys[i], objv)
		datas[i], err = rc.objMarshaller.Marshal(objv)
		if err != nil {
			return err
		}
	}
	return rc.BatchSet(ctx, keys, datas, expiration)
}

func (rc *redisClient) HSetObjects(ctx context.Context, key string, values ...any) error {
	if len(values) == 1 {
		value := values[0]
		switch v := value.(type) {
		case []any:
			values = v
		case map[string]any:
			values = make([]any, 0, 2*len(v))
			for kk, vv := range v {
				values = append(values, kk, vv)
			}
		default:
			panic(PanicHSetUnsupportedValueType)
		}
	}

	var l = make([]any, 0, len(values))
	if len(values)%2 != 0 {
		panic(PanicFieldValueCountUnmatched)
	}
	for i := 0; i < len(values); i += 2 {
		l = append(l, values[i]) // key
		bys, err := rc.objMarshaller.Marshal(values[i+1])
		if err != nil {
			return err
		}
		l = append(l, bys) // value
	}

	return rc.HSet(ctx, key, l...)
}

func (rc *redisClient) HGetObject(ctx context.Context, key string, field string, obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	data, err := rc.HGet(ctx, key, field)
	if err != nil {
		return err
	}
	err = rc.objMarshaller.Unmarshal([]byte(data), obj)
	return err
}

func (rc *redisClient) HMGetObjects(ctx context.Context, key string, fields []string, objs any) error {
	if len(fields) == 0 {
		panic(PanicFieldsIsMissing)
	}
	//start := time.Now().UnixMilli()
	objsValue := reflect.ValueOf(objs)
	if objsValue.Kind() != reflect.Slice {
		panic(PanicValueNeedBeSlice)
	}
	if len(fields) != objsValue.Len() {
		panic(PanicFieldValueCountUnmatched)
	}

	var elemIsInterface bool //objs type is []interface, the element in the slice may have different real type
	for i := 0; i < len(fields); i++ {
		if objsValue.Index(i).Kind() != reflect.Ptr && // input is slice of known struct type
			objsValue.Index(i).Elem().Kind() != reflect.Ptr { // input is slice of interface
			panic(PanicValueDstNeedBePointer)
		}
	}
	if objsValue.Index(0).Kind() != reflect.Ptr {
		elemIsInterface = true
	}

	rets, err := rc.HMGet(ctx, key, fields...)
	if err != nil {
		return err
	}
	//start1 := time.Now().UnixMilli()
	for i, v := range rets {
		objv := objsValue.Index(i)
		str, ok := v.(string)
		if !ok || str == "" {
			if !elemIsInterface {
				objv.Set(reflect.Zero(objv.Type()))
			} else {
				objv.Set(reflect.Zero(objv.Elem().Type()))
			}
			continue
		}

		if !elemIsInterface {
			if objv.IsNil() {
				objv.Set(reflect.New(objv.Type().Elem()))
			}
		} else {
			if objv.Elem().IsNil() {
				objv.Set(reflect.New(objv.Elem().Type().Elem()))
			}
		}
		//start := time.Now().UnixMilli()
		err = rc.objMarshaller.Unmarshal([]byte(str), objv.Interface())
		if err != nil {
			return err
		}
		//fmt.Println("HMGetObjects unmarshal", key, time.Now().UnixMilli()-start, "ms ", len(str))
	}
	//end := time.Now().UnixMilli()
	//fmt.Println("HMGetObjects", key, end-start, "ms ", end-start1, "ms")
	return nil
}

func (rc *redisClient) HGetAllObjects(ctx context.Context, key string, fields *[]string, objs any) error {
	if fields == nil {
		panic("HGetAllObjects fields is nil")
	}
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.HGetAll(ctx, key)
	if err != nil {
		return err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	var i = 0
	for k, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v), obj.Interface())
		if err != nil {
			return err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
		*fields = append(*fields, k)
		i++
	}
	return nil
}

func (rc *redisClient) ZAddObjects(ctx context.Context, key string, values ...any) (int64, error) {
	var l = make([]any, 0, len(values))
	if len(values)%2 != 0 {
		panic(PanicScoreValueCountUnmatched)
	}
	for i := 0; i < len(values); i += 2 {
		l = append(l, values[i]) // key
		bys, err := rc.objMarshaller.Marshal(values[i+1])
		if err != nil {
			return 0, err
		}
		l = append(l, bys) // value
	}
	return rc.ZAdd(ctx, key, l...)
}

func (rc *redisClient) ZRangeObjects(ctx context.Context, key string, start, stop int64, objs any) error {
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.ZRange(ctx, key, start, stop)
	if err != nil {
		return err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	for i, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v), obj.Interface())
		if err != nil {
			return err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
	}
	return nil
}

func (rc *redisClient) ZRangeObjectsByScore(ctx context.Context, key string, min, max string, objs any) error {
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.ZRangeByScore(ctx, key, min, max)
	if err != nil {
		return err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	for i, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v), obj.Interface())
		if err != nil {
			return err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
	}
	return nil
}

func (rc *redisClient) ZRangeObjectsWithScores(ctx context.Context, key string, start, stop int64, objs any) (
	scores []float64, err error) {
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.ZRangeWithScores(ctx, key, start, stop)
	if err != nil {
		return nil, err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	scores = make([]float64, 0, len(rets))
	for i, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v.Member.(string)), obj.Interface())
		if err != nil {
			return nil, err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
		scores = append(scores, v.Score)
	}
	return scores, nil
}

func (rc *redisClient) ZRangeObjectsByScoreWithScores(ctx context.Context, key string, min, max string, objs any) (
	scores []float64, err error) {
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.ZRangeByScoreWithScores(ctx, key, min, max)
	if err != nil {
		return nil, err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	scores = make([]float64, 0, len(rets))
	for i, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v.Member.(string)), obj.Interface())
		if err != nil {
			return nil, err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
		scores = append(scores, v.Score)
	}
	return scores, nil
}

func (rc *redisClient) ZRevRangeObjects(ctx context.Context, key string, start, stop int64, objs any) error {
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.ZRevRange(ctx, key, start, stop)
	if err != nil {
		return err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	for i, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v), obj.Interface())
		if err != nil {
			return err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
	}
	return nil
}

func (rc *redisClient) ZRevRangeObjectsByScore(ctx context.Context, key string, min, max string, objs any) error {
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.ZRevRangeByScore(ctx, key, min, max)
	if err != nil {
		return err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	for i, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v), obj.Interface())
		if err != nil {
			return err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
	}
	return nil
}

func (rc *redisClient) ZRevRangeObjectsWithScores(ctx context.Context, key string, start, stop int64, objs any) (
	scores []float64, err error) {
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.ZRevRangeWithScores(ctx, key, start, stop)
	if err != nil {
		return nil, err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	scores = make([]float64, 0, len(rets))
	for i, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v.Member.(string)), obj.Interface())
		if err != nil {
			return nil, err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
		scores = append(scores, v.Score)
	}
	return scores, nil
}

func (rc *redisClient) ZRevRangeObjectsByScoreWithScores(ctx context.Context, key string, min, max string, objs any) (
	scores []float64, err error) {
	// check type of objs
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr {
		panic(PanicValueDstNeedBePointer)
	}
	// get the elem type
	objsValue := reflect.ValueOf(objs)
	objType := objsValue.Elem().Type().Elem()
	var objPtr = objType.Kind() == reflect.Ptr

	rets, err := rc.ZRevRangeByScoreWithScores(ctx, key, min, max)
	if err != nil {
		return nil, err
	}

	objsValue.Elem().Set(reflect.MakeSlice(objsValue.Elem().Type(), len(rets), len(rets)))
	scores = make([]float64, 0, len(rets))
	for i, v := range rets {
		t := objType
		if objPtr {
			t = objType.Elem()
		}
		obj := reflect.New(t)
		err = rc.objMarshaller.Unmarshal([]byte(v.Member.(string)), obj.Interface())
		if err != nil {
			return nil, err
		}
		if objPtr {
			objsValue.Elem().Index(i).Set(obj)
		} else {
			objsValue.Elem().Index(i).Set(obj.Elem())
		}
		scores = append(scores, v.Score)
	}
	return scores, nil
}

func (rc *redisClient) ZRankObject(ctx context.Context, key string, member any) (int64, error) {
	data, err := rc.objMarshaller.Marshal(member)
	if err != nil {
		return 0, err
	}
	return rc.ZRank(ctx, key, string(data))
}

func (rc *redisClient) ZRevRankObject(ctx context.Context, key string, member any) (int64, error) {
	data, err := rc.objMarshaller.Marshal(member)
	if err != nil {
		return 0, err
	}
	return rc.ZRevRank(ctx, key, string(data))
}

func (rc *redisClient) ZScoreObject(ctx context.Context, key string, member any) (float64, error) {
	data, err := rc.objMarshaller.Marshal(member)
	if err != nil {
		return 0, err
	}
	return rc.ZScore(ctx, key, string(data))
}

func (rc *redisClient) ZRemObjects(ctx context.Context, key string, members ...any) (int64, error) {
	for i, v := range members {
		bys, err := rc.objMarshaller.Marshal(v)
		if err != nil {
			return 0, err
		}
		members[i] = string(bys)
	}
	return rc.ZRem(ctx, key, members...)
}
