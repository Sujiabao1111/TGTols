package helpers

// for fun
// func ZeroValue(value *int32) {
// 	*value ^= *value
// }

// func NegateValue(value *int32) int32 {
// 	return (*value ^ (-1)) + 1
// }

// func ConditionallyNegateValue(value uint32, flag int32) uint32 {
// 	return (uint32)(flag^(flag-1)) * (value)
// }
// for fun end

// "d" needs to be the power of two
func ModValue(value, d uint32) uint32 {
	return value & (d - 1)
}

func AbsValue(value int32) int32 {
	return value + (value >> 31) ^ (value >> 31)
}

func HasZeroByte(x int) bool {
	if x&0x000000FF == 0 {
		return true
	}
	if x&0x0000FF00 == 0 {
		return true
	}
	if x&0x00FF0000 == 0 {
		return true
	}
	if x&0xFF000000 == 0 {
		return true
	}
	return false
}

func CountEvenBranchFreeNoDeps(numbers []int) int {
	c1 := 0
	c2 := 0
	c3 := 0
	c4 := 0
	l := len(numbers)
	for i := 0; i < l; i += 4 {
		c1 += numbers[i] & 1
		if i+1 < l {
			c2 += numbers[i+1] & 1
		}
		if i+2 < l {
			c3 += numbers[i+2] & 1
		}
		if i+3 < l {
			c4 += numbers[i+3] & 1
		}
	}

	return l - c1 - c2 - c3 - c4
}
