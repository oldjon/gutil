package conv

import "strconv"

// StringToUInt32 convert string to uint32
func StringToUInt32(v string) uint32 {
	i, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(i)
}

// StringToBool convert string to bool
func StringToBool(v string) bool {
	i, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}
	return i
}

// UInt32ToString convert uint32 to string
func UInt32ToString(v uint32) string {
	return strconv.FormatUint(uint64(v), 10)
}

// StringToInt32 convert string to int32
func StringToInt32(v string) int32 {
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0
	}
	return int32(i)
}

// UIn64ToString convert uint64 to string
func UIn64ToString(v uint64) string {
	return strconv.FormatUint(v, 10)
}

// StringToUInt64 convert string to uint64
func StringToUInt64(v string) uint64 {
	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return uint64(i)
}

// StringToFloat32 convert string to float32
func StringToFloat32(v string) float32 {
	i, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return 0
	}
	return float32(i)
}

// StringToFloat64 convert string to float64
func StringToFloat64(v string) float64 {
	i, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return i
}

// StringToInt64 convert string to int64
func StringToInt64(v string) int64 {
	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return int64(i)
}

// HexStringToInt64 convert string to int64
func HexStringToInt64(v string) int64 {
	i, err := strconv.ParseUint(v, 0, 64)
	if err != nil {
		return 0
	}
	return int64(i)
}
