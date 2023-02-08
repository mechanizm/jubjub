package fr

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/mechanizm/jubjub/futil"
)

type Fr [4]uint64

func FromBytes(byt []byte) *Fr {
	d := &Fr{0, 0, 0, 0}

	d[0] = binary.LittleEndian.Uint64(byt[0:8])
	d[1] = binary.LittleEndian.Uint64(byt[8:16])
	d[2] = binary.LittleEndian.Uint64(byt[16:24])
	d[3] = binary.LittleEndian.Uint64(byt[24:32])

	// Convert to Montgomery form
	return d.Mul(&R2)
}

func FromBytesWide(byt []byte) *Fr {
	d0 := &Fr{0, 0, 0, 0}
	d1 := &Fr{0, 0, 0, 0}

	d0[0] = binary.LittleEndian.Uint64(byt[0:8])
	d0[1] = binary.LittleEndian.Uint64(byt[8:16])
	d0[2] = binary.LittleEndian.Uint64(byt[16:24])
	d0[3] = binary.LittleEndian.Uint64(byt[24:32])

	d1[0] = binary.LittleEndian.Uint64(byt[32:40])
	d1[1] = binary.LittleEndian.Uint64(byt[40:48])
	d1[2] = binary.LittleEndian.Uint64(byt[48:56])
	d1[3] = binary.LittleEndian.Uint64(byt[56:64])

	// Convert to Montgomery form
	d0 = d0.Mul(&R2)
	d1 = d1.Mul(&R3)

	return d0.Add(d1)
}

func Zero() *Fr {
	return &Fr{0, 0, 0, 0}
}

func One() *Fr {
	return &R
}

func (lhs *Fr) Add(rhs *Fr) *Fr {
	d0, carry := futil.Adc(lhs[0], rhs[0], 0)
	d1, carry := futil.Adc(lhs[1], rhs[1], carry)
	d2, carry := futil.Adc(lhs[2], rhs[2], carry)
	d3, _ := futil.Adc(lhs[3], rhs[3], carry)

	f := &Fr{0, 0, 0, 0}
	f[0] = d0
	f[1] = d1
	f[2] = d2
	f[3] = d3
	// Normalise
	return f.Sub(&r)
}

// Sub Subtracts one field from another
func (lhs *Fr) Sub(rhs *Fr) *Fr {
	d0, borrow := futil.Sbb(lhs[0], rhs[0], 0)
	d1, borrow := futil.Sbb(lhs[1], rhs[1], borrow)
	d2, borrow := futil.Sbb(lhs[2], rhs[2], borrow)
	d3, borrow := futil.Sbb(lhs[3], rhs[3], borrow)

	// If underflow occurred on the final limb, borrow = 0xfff...fff, otherwise
	// // borrow = 0x000...000. Thus, we use it as a mask to conditionally add the modulus.
	d0, carry := futil.Adc(d0, r[0]&borrow, 0)
	d1, carry = futil.Adc(d1, r[1]&borrow, carry)
	d2, carry = futil.Adc(d2, r[2]&borrow, carry)
	d3, carry = futil.Adc(d3, r[3]&borrow, carry)

	f := &Fr{0, 0, 0, 0}
	f[0] = d0
	f[1] = d1
	f[2] = d2
	f[3] = d3

	return f
}

func (f *Fr) Neg() *Fr {
	d0, borrow := futil.Sbb(MODULUS[0], f[0], 0)
	d1, borrow := futil.Sbb(MODULUS[1], f[1], borrow)
	d2, borrow := futil.Sbb(MODULUS[2], f[2], borrow)
	d3, _ := futil.Sbb(MODULUS[3], f[3], borrow)

	msk := f[0]|f[1]|f[2]|f[3] == 0
	var mask uint64
	if !msk {
		mask--
	}

	ff := &Fr{0, 0, 0, 0}
	ff[0] = d0 & mask
	ff[1] = d1 & mask
	ff[2] = d2 & mask
	ff[3] = d3 & mask
	return ff
}

// Mul mutiplies two field elements together
func (lhs *Fr) Mul(rhs *Fr) *Fr {
	r0, carry := futil.Mac(0, lhs[0], rhs[0], 0)
	r1, carry := futil.Mac(0, lhs[0], rhs[1], carry)
	r2, carry := futil.Mac(0, lhs[0], rhs[2], carry)
	r3, r4 := futil.Mac(0, lhs[0], rhs[3], carry)

	r1, carry = futil.Mac(r1, lhs[1], rhs[0], 0)
	r2, carry = futil.Mac(r2, lhs[1], rhs[1], carry)
	r3, carry = futil.Mac(r3, lhs[1], rhs[2], carry)
	r4, r5 := futil.Mac(r4, lhs[1], rhs[3], carry)

	r2, carry = futil.Mac(r2, lhs[2], rhs[0], 0)
	r3, carry = futil.Mac(r3, lhs[2], rhs[1], carry)
	r4, carry = futil.Mac(r4, lhs[2], rhs[2], carry)
	r5, r6 := futil.Mac(r5, lhs[2], rhs[3], carry)

	r3, carry = futil.Mac(r3, lhs[3], rhs[0], 0)
	r4, carry = futil.Mac(r4, lhs[3], rhs[1], carry)
	r5, carry = futil.Mac(r5, lhs[3], rhs[2], carry)
	r6, r7 := futil.Mac(r6, lhs[3], rhs[3], carry)

	red := MontRed(r0, r1, r2, r3, r4, r5, r6, r7)

	f := &Fr{0, 0, 0, 0}
	f[0] = red[0]
	f[1] = red[1]
	f[2] = red[2]
	f[3] = red[3]
	return f
}

func MontRed(r0, r1, r2, r3, r4, r5, r6, r7 uint64) *Fr {
	k := r0 * INV
	_, carry := futil.Mac(r0, k, r[0], 0)
	r1, carry = futil.Mac(r1, k, r[1], carry)
	r2, carry = futil.Mac(r2, k, r[2], carry)
	r3, carry = futil.Mac(r3, k, r[3], carry)
	r4, carry2 := futil.Adc(r4, 0, carry)

	k = r1 * INV
	_, carry = futil.Mac(r1, k, r[0], 0)
	r2, carry = futil.Mac(r2, k, r[1], carry)
	r3, carry = futil.Mac(r3, k, r[2], carry)
	r4, carry = futil.Mac(r4, k, r[3], carry)
	r5, carry2 = futil.Adc(r5, carry2, carry)

	k = r2 * INV
	_, carry = futil.Mac(r2, k, r[0], 0)
	r3, carry = futil.Mac(r3, k, r[1], carry)
	r4, carry = futil.Mac(r4, k, r[2], carry)
	r5, carry = futil.Mac(r5, k, r[3], carry)
	r6, carry2 = futil.Adc(r6, carry2, carry)

	k = r3 * INV
	_, carry = futil.Mac(r3, k, r[0], 0)
	r4, carry = futil.Mac(r4, k, r[1], carry)
	r5, carry = futil.Mac(r5, k, r[2], carry)
	r6, carry = futil.Mac(r6, k, r[3], carry)
	r7, carry2 = futil.Adc(r7, carry2, carry)

	f := &Fr{r4, r5, r6, r7}

	return f.Sub(&r)
}

func (f *Fr) Double() *Fr {
	return f.Add(f)
}

// IntoBytes  converts f into a little endian byte slice
func (f *Fr) Bytes() []byte {
	// Turn into canonical form by computing (a.R) / R = a
	tmp := MontRed(f[0], f[1], f[2], f[3], 0, 0, 0, 0)

	res := make([]byte, 32, 32)

	binary.LittleEndian.PutUint64(res[0:8], tmp[0])
	binary.LittleEndian.PutUint64(res[8:16], tmp[1])
	binary.LittleEndian.PutUint64(res[16:24], tmp[2])
	binary.LittleEndian.PutUint64(res[24:32], tmp[3])

	return res
}

func (f *Fr) String() string {
	s := f.Bytes()

	// reverse bytes
	for i, j := 0, len(s)-1; i <= j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return hex.EncodeToString(s[:])
}
