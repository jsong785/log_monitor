package core_utils

func ReverseBytes(slice []byte) []byte {
	if len(slice) == 0 {
		return slice
	}
	start := 0
	end := len(slice) - 1
	for start < end {
		a := slice[start]
		b := slice[end]

		slice[start] = b
		slice[end] = a
		start++
		end--
	}
	return slice
}
