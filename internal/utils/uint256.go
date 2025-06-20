package utils

type Uint256 struct {
	hi, m1, m2, lo uint64
}

func (u Uint256) And(a Uint256) Uint256 {
	return Uint256{a.hi & u.hi, u.m1 & a.m1, u.m2 & a.m2, a.lo & u.lo}
}

func (u Uint256) Xor(a Uint256) Uint256 {
	return Uint256{a.hi ^ u.hi, u.m1 ^ a.m1, u.m2 ^ a.m2, a.lo ^ u.lo}
}

func (u Uint256) Or(a Uint256) Uint256 {
	return Uint256{a.hi | u.hi, u.m1 | a.m1, u.m2 | a.m2, a.lo | u.lo}
}
