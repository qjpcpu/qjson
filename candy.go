package qjson

import "unsafe"

type tinyMap []byte

/* convert map to array for better performance */
func mapToSlice(m map[byte]bool) tinyMap {
	var max byte
	for b := range m {
		if b > max {
			max = b
		}
	}
	arr := make([]byte, int(max+1), int(max+1))
	for b := range m {
		arr[int(b)] = max
	}
	return arr
}

func (m tinyMap) Contains(b byte) bool {
	return int(b) < len(m) && m[int(b)] != 0
}

func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func isIntegerChar(b byte) bool {
	return b >= '0' && b <= '9'
}
