package mos65xx

// adc calculation
func adc(a, b uint8, carry, bcd bool) (r uint8, n, v, z, c bool) {
	t := uint16(a) + uint16(b)
	if carry {
		t++
	}

	if bcd {
		lo := (a & 0x0f) + (b & 0x0f)
		if carry {
			lo++
		}
		if lo > 0x09 {
			t += 0x06
		}
		if t > 0x99 {
			t += 0x60
		}
		carry = t > 0x99
	} else {
		carry = t&0xff00 != 0
	}

	r = uint8(t)
	n = r&0x80 == 0x80
	v = overflow(a, b, r)
	z = r == 0
	c = carry
	return
}

// sbc calculation
func sbc(a, b uint8, carry, bcd bool) (r uint8, n, v, z, c bool) {
	t := uint16(a) - uint16(b)
	if !carry {
		t--
	}

	if bcd {
		lo := (a & 0x0f) - (b & 0x0f)
		if !carry {
			lo--
		}
		if lo&0xf0 != 0 {
			t -= 0x06
		}
		if t > 0x99 {
			t -= 0x60
		}
	}

	r = uint8(t)
	n = r&0x80 == 0x80
	v = underflow(a, b, r)
	z = r == 0
	c = t < 0x100
	return
}
