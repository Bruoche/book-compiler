package chars

func Range(first, last byte) []byte {
	var chars []byte
	i := first
	for i <= last {
		chars = append(chars, i)
		i++
	}
	return chars
}

var Uncapped = Range('a', 'z')
var Capitalized = Range('A', 'Z')
