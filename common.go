package fortuna

func bytesToInt64(bytes []byte) int64 {
	var res int64
	res = int64(bytes[0])
	for _, x := range bytes[1:] {
		res = res<<8 | int64(x)
	}
	return res
}

func int64ToBytes(x int64) []byte {
	bytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		bytes[i] = byte(x & 0xff)
		x = x >> 8
	}
	return bytes
}

func isZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}
