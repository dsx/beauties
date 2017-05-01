package beauties

const (
	symbols = "23456789abcdefghjkmnpqrstwxyz"
	base    = int64(len(symbols))
)

// Encode encodes a symbol using symbol base
// Taken from https://github.com/fs111/kurz.go/blob/master/src/codec.go
func Encode(number int64) (result string) {
	rest := number % base
	result = string(symbols[rest])
	if number-rest != 0 {
		n := (number - rest) / base
		result = Encode(n) + result
	}
	return
}
