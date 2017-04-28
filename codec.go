package beauties

const (
	symbols = "23456789abcdefghjkmnpqrstwxyz"
	base    = int64(len(symbols))
)

// Encode encodes a number into our *base* representation
// Taken from https://github.com/fs111/kurz.go/blob/master/src/codec.go
func Encode(number int64) string {
	rest := number % base
	result := string(symbols[rest])
	if number-rest != 0 {
		newnumber := (number - rest) / base
		result = Encode(newnumber) + result
	}
	return result
}
