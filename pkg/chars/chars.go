package chars

func Range(first, last byte) []byte {
	var chars []byte
	i := first
	for i <= last {
		if len(chars) > 0 && i == 0 {
			return chars // Infinite loop! (chars loop to 0 when overflowing)
		}
		chars = append(chars, i)
		i++
	}
	return chars
}

var Uncapped = Range('a', 'z')
var Capitalized = Range('A', 'Z')
var Letters = append(Uncapped, Capitalized...)
var All = Range(0, 255)
var Nums = Range('1', '9')
