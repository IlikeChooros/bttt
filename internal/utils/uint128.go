package utils

type Uint128 struct {
	high uint64
	low  uint64
}

func (u Uint128) And(a Uint128) Uint128 {
	return Uint128{a.high & u.high, a.low & u.low}
}

func (u Uint128) Xor(a Uint128) Uint128 {
	return Uint128{a.high ^ u.high, a.low ^ u.low}
}

func (u Uint128) Or(a Uint128) Uint128 {
	return Uint128{u.high | a.high, u.low | a.low}
}
