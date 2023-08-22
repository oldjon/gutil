package gdb

import (
	"strconv"

	"github.com/go-redis/redis/v8"
)

func ToUint64(v any, err error) (uint64, error) {
	if err != nil {
		return 0, err
	}
	return toUint64(v)
}

func toUint64(v any) (uint64, error) {
	switch vv := v.(type) {
	case string:
		vStr := v.(string)
		n, err := strconv.Atoi(vStr)
		if err != nil {
			return 0, ErrValue
		}
		return uint64(n), nil
	case int:
		return uint64(vv), nil
	case int8:
		return uint64(vv), nil
	case int16:
		return uint64(vv), nil
	case int32:
		return uint64(vv), nil
	case int64:
		return uint64(vv), nil
	case uint:
		return uint64(vv), nil
	case uint8:
		return uint64(vv), nil
	case uint16:
		return uint64(vv), nil
	case uint32:
		return uint64(vv), nil
	case uint64:
		return vv, nil
	case float32:
		return uint64(vv), nil
	case float64:
		return uint64(vv), nil
	case *redis.IntCmd:
		vi := v.(*redis.IntCmd)
		return uint64(vi.Val()), nil
	case *redis.FloatCmd:
		vf := v.(*redis.FloatCmd)
		return uint64(vf.Val()), nil
	default:
		return 0, ErrValueType
	}
}

func ToUint32(v any, err error) (uint32, error) {
	if err != nil {
		return 0, err
	}
	return toUint32(v)
}

func toUint32(v any) (uint32, error) {
	switch vv := v.(type) {
	case string:
		vStr := v.(string)
		n, err := strconv.Atoi(vStr)
		if err != nil {
			return 0, ErrValue
		}
		return uint32(n), nil
	case int:
		return uint32(vv), nil
	case int8:
		return uint32(vv), nil
	case int16:
		return uint32(vv), nil
	case int32:
		return uint32(vv), nil
	case int64:
		return uint32(vv), nil
	case uint:
		return uint32(vv), nil
	case uint8:
		return uint32(vv), nil
	case uint16:
		return uint32(vv), nil
	case uint32:
		return vv, nil
	case uint64:
		return uint32(vv), nil
	case float32:
		return uint32(vv), nil
	case float64:
		return uint32(vv), nil
	case *redis.IntCmd:
		vi := v.(*redis.IntCmd)
		return uint32(vi.Val()), nil
	case *redis.FloatCmd:
		vf := v.(*redis.FloatCmd)
		return uint32(vf.Val()), nil
	default:
		return 0, ErrValueType
	}
}

func ToString(v any, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return toString(v)
}

func toString(v any) (string, error) {
	switch vv := v.(type) {
	case int:
		return strconv.Itoa(vv), nil
	case int8:
		return strconv.Itoa(int(vv)), nil
	case int32:
		return strconv.Itoa(int(vv)), nil
	case int64:
		return strconv.Itoa(int(vv)), nil
	case uint:
		return strconv.Itoa(int(vv)), nil
	case uint8:
		return strconv.Itoa(int(vv)), nil
	case uint16:
		return strconv.Itoa(int(vv)), nil
	case uint32:
		return strconv.Itoa(int(vv)), nil
	case uint64:
		return strconv.Itoa(int(vv)), nil
	case []byte:
		return string(vv), nil
	case string:
		return vv, nil
	case *redis.IntCmd:
		return strconv.Itoa(int(v.(*redis.IntCmd).Val())), nil
	case *redis.StringCmd:
		return v.(*redis.StringCmd).Val(), nil
	default:
		return "", ErrValueType
	}
}

func toFloat64(v any) (float64, error) {
	switch vv := v.(type) {
	case int:
		return float64(vv), nil
	case int8:
		return float64(vv), nil
	case int32:
		return float64(vv), nil
	case int64:
		return float64(vv), nil
	case uint:
		return float64(vv), nil
	case uint8:
		return float64(vv), nil
	case uint16:
		return float64(vv), nil
	case uint32:
		return float64(vv), nil
	case uint64:
		return float64(vv), nil
	case float32:
		return float64(vv), nil
	case float64:
		return vv, nil
	default:
		return 0, ErrValueType
	}
}

// ToStringSlice convert []*redis.StringCmd, *redis.SliceCmd and *redis.StringSliceCmd to []string,
// when *redis.Slice, the value inside should be string.
func ToStringSlice(values any, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	switch values.(type) {
	case []*redis.StringCmd:
		rv := values.([]*redis.StringCmd)
		strs := make([]string, len(rv))
		for i, v := range rv {
			strs[i] = v.Val()
		}
		return strs, nil
	case *redis.SliceCmd:
		rv := values.(*redis.SliceCmd).Val()
		strs := make([]string, len(rv))
		for i, v := range rv {
			str, ok := v.(string)
			if !ok {
				return nil, ErrValueType
			}
			strs[i] = str
		}
		return strs, nil
	case *redis.StringSliceCmd:
		return values.(*redis.StringSliceCmd).Val(), nil
	default:
		return nil, ErrValueType
	}
}

// ToStringStringMap convert *redis.StringStringMapCmd
// to []string, when *redis.Slice, the value inside should be string.
func ToStringStringMap(values any, err error) (map[string]string, error) {
	if err != nil {
		return nil, err
	}
	switch vv := values.(type) {
	case *redis.StringStringMapCmd:
		return vv.Val(), nil
	default:
		return nil, ErrValueType
	}
}

func ToUint64Slice(values any, err error) ([]uint64, error) {
	if err != nil {
		return nil, err
	}
	switch values.(type) {
	case []redis.Z:
		rv := values.([]redis.Z)
		sl := make([]uint64, 0, len(rv))
		for _, v := range rv {
			m, err := toUint64(v.Member)
			if err != nil {
				panic(PanicValueNotNum)
			}
			sl = append(sl, m, uint64(v.Score))
		}
		return sl, nil
	default:
		return nil, ErrValueType
	}
}
