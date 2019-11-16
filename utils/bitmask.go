package utils

type Bitmask uint32

func (f Bitmask) Has(flag Bitmask) bool {
	return f&flag != 0
}
func (f *Bitmask) Set(flag Bitmask) {
	*f |= flag
}
func (f *Bitmask) Unset(flag Bitmask) {
	*f &= ^flag
}
func (f *Bitmask) Toggle(flag Bitmask) {
	*f ^= flag
}
