package fq

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/jadeydi/jubjub/pkg/jubjub/futil"
)

type Fq [4]uint64

//from_bytes
func FromBytes(byt []byte) *Fq {
	d := &Fq{0, 0, 0, 0}

	d[0] = binary.LittleEndian.Uint64(byt[0:8])
	d[1] = binary.LittleEndian.Uint64(byt[8:16])
	d[2] = binary.LittleEndian.Uint64(byt[16:24])
	d[3] = binary.LittleEndian.Uint64(byt[24:32])

	// Convert to Montgomery form
	d.Mul(d, &R2)

	return d
}

// Sub Subtracts one field from another
func (a *Fq) Sub(b *Fq) *Fq {
	var borrow, carry uint64
	d0, borrow := futil.Sbb(a[0], b[0], borrow)
	d1, borrow := futil.Sbb(a[1], b[1], borrow)
	d2, borrow := futil.Sbb(a[2], b[2], borrow)
	d3, borrow := futil.Sbb(a[3], b[3], borrow)

	// If underflow occurred on the final limb, borrow = 0xfff...fff, otherwise
	// borrow = 0x000...000. Thus, we use it as a mask to conditionally add the modulus.
	d0, carry = futil.Adc(d0, q[0]&borrow, carry)
	d1, carry = futil.Adc(d1, q[1]&borrow, carry)
	d2, carry = futil.Adc(d2, q[2]&borrow, carry)
	d3, carry = futil.Adc(d3, q[3]&borrow, carry)

	f := &Fq{0, 0, 0, 0}
	f[0] = d0
	f[1] = d1
	f[2] = d2
	f[3] = d3

	return f
}

// Neg negates a Fq
func (a *Fq) Neg() *Fq {
	d0, borrow := futil.Sbb(q[0], a[0], 0)
	d1, borrow := futil.Sbb(q[1], a[1], borrow)
	d2, borrow := futil.Sbb(q[2], a[2], borrow)
	d3, _ := futil.Sbb(q[3], a[3], borrow)

	msk := a[0]|a[1]|a[2]|a[3] == 0

	var mask uint64
	if !msk {
		mask-- // uint64 max
	}

	// `tmp` could be `MODULUS` if `self` was zero. Create a mask that is
	// zero if `self` was zero, and `u64::max_value()` if self was nonzero.
	f := &Fq{0, 0, 0, 0}
	f[0] = d0 & mask
	f[1] = d1 & mask
	f[2] = d2 & mask
	f[3] = d3 & mask

	return f
}

// Add Adds one field to another
func (lhs *Fq) Add(rhs *Fq) *Fq {
	d0, carry := futil.Adc(lhs[0], rhs[0], 0)
	d1, carry := futil.Adc(lhs[1], rhs[1], carry)
	d2, carry := futil.Adc(lhs[2], rhs[2], carry)
	d3, _ := futil.Adc(lhs[3], rhs[3], carry)

	f := &Fq{0, 0, 0, 0}
	f[0] = d0
	f[1] = d1
	f[2] = d2
	f[3] = d3

	return f.Sub(&q)
}

func (f *Fq) Mul(lhs, rhs *Fq) *Fq {
	// TODO: Optimise later
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

	*f = *montRed(r0, r1, r2, r3, r4, r5, r6, r7)

	return f
}

// Double doubles f by adding it to itself
func (f *Fq) Double() *Fq {
	return f.Add(f)
}

// Equal returns true, if a ==b
func (a *Fq) Equal(b *Fq) bool {
	return a[0] == b[0] && a[1] == b[1] && a[2] == b[2] && a[3] == b[3]
}

func (f *Fq) Square(a *Fq) *Fq {
	r1, carry := futil.Mac(0, a[0], a[1], 0)
	r2, carry := futil.Mac(0, a[0], a[2], carry)
	r3, r4 := futil.Mac(0, a[0], a[3], carry)

	r3, carry = futil.Mac(r3, a[1], a[2], 0)
	r4, r5 := futil.Mac(r4, a[1], a[3], carry)

	r5, r6 := futil.Mac(r5, a[2], a[3], 0)

	r7 := r6 >> 63
	r6 = (r6 << 1) | (r5 >> 63)
	r5 = (r5 << 1) | (r4 >> 63)
	r4 = (r4 << 1) | (r3 >> 63)
	r3 = (r3 << 1) | (r2 >> 63)
	r2 = (r2 << 1) | (r1 >> 63)
	r1 = r1 << 1

	r0, carry := futil.Mac(0, a[0], a[0], 0)
	r1, carry = futil.Adc(0, r1, carry)
	r2, carry = futil.Mac(r2, a[1], a[1], carry)
	r3, carry = futil.Adc(0, r3, carry)
	r4, carry = futil.Mac(r4, a[2], a[2], carry)
	r5, carry = futil.Adc(0, r5, carry)

	r6, carry = futil.Mac(r6, a[3], a[3], carry)
	r7, _ = futil.Adc(0, r7, carry)

	red := montRed(r0, r1, r2, r3, r4, r5, r6, r7)
	f[0] = red[0]
	f[1] = red[1]
	f[2] = red[2]
	f[3] = red[3]

	return f
}

func (f *Fq) SqrtVarTime() *Fq {
	one := One()
	zero := &Fq{0, 0, 0, 0}
	tmp := &Fq{0, 0, 0, 0}

	*tmp = *f

	lgs := tmp.LegendreSymbolVarTime()

	if lgs.Equal(zero) {
		return f
	}
	if !lgs.Equal(one) {
		f = nil
		return nil // XXX: We should bubble up an error for this
	}

	*tmp = *f
	r := *tmp.PowVarTime([4]uint64{0x7fff2dff80000000, 0x04d0ec02a9ded201, 0x94cebea4199cec04, 0x0000000039f6d3a9})

	*tmp = *f
	t := *tmp.PowVarTime([4]uint64{0xfffe5bfeffffffff, 0x09a1d80553bda402, 0x299d7d483339d808, 0x0000000073eda753})

	c := ROOTOFUNITY
	m := S

	for !t.Equal(one) {

		var i = uint32(1)

		t2i := &Fq{0, 0, 0, 0}
		t2i.Square(&t)

		for !t2i.Equal(one) {
			t2i.Square(t2i)
			i++
		}

		for k := uint32(0); k < m-i-1; k++ {
			c.Square(&c)
		}

		r.Mul(&r, &c)
		c.Square(&c)
		t.Mul(&t, &c)
		m = i

	}

	*f = r

	return f
}

func (f *Fq) LegendreSymbolVarTime() *Fq {
	// Legendre symbol computed via Euler's criterion:
	// self^((q - 1) // 2)
	f.PowVarTime([4]uint64{
		0x7fffffff80000000,
		0xa9ded2017fff2dff,
		0x199cec0404d0ec02,
		0x39f6d3a994cebea4,
	})
	return f
}

func (f *Fq) PowVarTime(b [4]uint64) *Fq {
	res := One()

	for j := range b {

		e := b[len(b)-1-j] // reversed
		for i := 63; i >= 0; i-- {

			res.Square(res)

			if ((e >> uint64(i)) & 1) == 1 {
				res.Mul(res, f)
			}

		}

	}
	*f = *res
	return f
}

// Inverse inverts a field element
// If element is zero, it will return nil
func (f *Fq) Inverse(a *Fq) *Fq {

	zero := Zero()

	// Check if f is non-zero
	if f.Equal(zero) {
		return nil
	}

	var sqrMulti = func(e *Fq, n uint64) {
		for i := uint64(0); i < n; i++ {
			e.Square(e)
		}
	}

	var t0, t1, t2, t3, t4, t5, t6, t7, t8, t9, t10, t11, t12, t13, t14, t15, t16, t17 Fq

	t10 = *a
	t0.Square(&t10)
	t1.Mul(&t0, &t10)
	t16.Square(&t0)
	t6.Square(&t16)
	t5.Mul(&t6, &t0)
	t0.Mul(&t6, &t16)
	t12.Mul(&t5, &t16)
	t2.Square(&t6)
	t7.Mul(&t5, &t6)
	t15.Mul(&t0, &t5)
	t17.Square(&t12)
	t1.Mul(&t1, &t17)
	t3.Mul(&t7, &t2)
	t8.Mul(&t1, &t17)
	t4.Mul(&t8, &t2)
	t9.Mul(&t8, &t7)
	t7.Mul(&t4, &t5)
	t11.Mul(&t4, &t17)
	t5.Mul(&t9, &t17)
	t14.Mul(&t7, &t15)
	t13.Mul(&t11, &t12)
	t12.Mul(&t11, &t17)
	t15.Mul(&t15, &t12)
	t16.Mul(&t16, &t15)
	t3.Mul(&t3, &t16)
	t17.Mul(&t17, &t3)
	t0.Mul(&t0, &t17)
	t6.Mul(&t6, &t0)
	t2.Mul(&t2, &t6)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t17)
	sqrMulti(&t0, 9)
	t0.Mul(&t0, &t16)
	sqrMulti(&t0, 9)
	t0.Mul(&t0, &t15)
	sqrMulti(&t0, 9)
	t0.Mul(&t0, &t15)
	sqrMulti(&t0, 7)
	t0.Mul(&t0, &t14)
	sqrMulti(&t0, 7)
	t0.Mul(&t0, &t13)
	sqrMulti(&t0, 10)
	t0.Mul(&t0, &t12)
	sqrMulti(&t0, 9)
	t0.Mul(&t0, &t11)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t8)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t10)
	sqrMulti(&t0, 14)
	t0.Mul(&t0, &t9)
	sqrMulti(&t0, 10)
	t0.Mul(&t0, &t8)
	sqrMulti(&t0, 15)
	t0.Mul(&t0, &t7)
	sqrMulti(&t0, 10)
	t0.Mul(&t0, &t6)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t5)
	sqrMulti(&t0, 16)
	t0.Mul(&t0, &t3)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t2)
	sqrMulti(&t0, 7)
	t0.Mul(&t0, &t4)
	sqrMulti(&t0, 9)
	t0.Mul(&t0, &t2)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t3)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t2)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t2)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t2)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t3)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t2)
	sqrMulti(&t0, 8)
	t0.Mul(&t0, &t2)
	sqrMulti(&t0, 5)
	t0.Mul(&t0, &t1)
	sqrMulti(&t0, 5)
	t0.Mul(&t0, &t1)

	f[0] = t0[0]
	f[1] = t0[1]
	f[2] = t0[2]
	f[3] = t0[3]
	return f
}

// Zero sets f to the zero element
func Zero() *Fq {
	var f Fq
	copy(f[:], zero[:])
	return &f
}

// One sets f to the one element
func One() *Fq {
	f := &Fq{1, 0, 0, 0}
	copy(f[:], R[:])
	return f
}

// BytesInto  converts f into a little endian byte slice
func (f *Fq) BytesInto(buf *[32]byte) *Fq {
	// Turn into canonical form by computing (a.R) / R = a
	tmp := *montRed(f[0], f[1], f[2], f[3], 0, 0, 0, 0)

	buf[0] = uint8(tmp[0])
	buf[1] = uint8(tmp[0] >> 8)
	buf[2] = uint8(tmp[0] >> 16)
	buf[3] = uint8(tmp[0] >> 24)
	buf[4] = uint8(tmp[0] >> 32)
	buf[5] = uint8(tmp[0] >> 40)
	buf[6] = uint8(tmp[0] >> 48)
	buf[7] = uint8(tmp[0] >> 56)
	buf[8] = uint8(tmp[1])
	buf[9] = uint8(tmp[1] >> 8)
	buf[10] = uint8(tmp[1] >> 16)
	buf[11] = uint8(tmp[1] >> 24)
	buf[12] = uint8(tmp[1] >> 32)
	buf[13] = uint8(tmp[1] >> 40)
	buf[14] = uint8(tmp[1] >> 48)
	buf[15] = uint8(tmp[1] >> 56)
	buf[16] = uint8(tmp[2])
	buf[17] = uint8(tmp[2] >> 8)
	buf[18] = uint8(tmp[2] >> 16)
	buf[19] = uint8(tmp[2] >> 24)
	buf[20] = uint8(tmp[2] >> 32)
	buf[21] = uint8(tmp[2] >> 40)
	buf[22] = uint8(tmp[2] >> 48)
	buf[23] = uint8(tmp[2] >> 56)
	buf[24] = uint8(tmp[3])
	buf[25] = uint8(tmp[3] >> 8)
	buf[26] = uint8(tmp[3] >> 16)
	buf[27] = uint8(tmp[3] >> 24)
	buf[28] = uint8(tmp[3] >> 32)
	buf[29] = uint8(tmp[3] >> 40)
	buf[30] = uint8(tmp[3] >> 48)
	buf[31] = uint8(tmp[3] >> 56)
	return f
}

func (f *Fq) SetBytes(b *[32]byte) *Fq {
	f[0] = futil.Load4(b[0:])
	f[1] = futil.Load4(b[8:])
	f[2] = futil.Load4(b[16:])
	f[3] = futil.Load4(b[24:]) & 0x1fffffff
	return f.Sub(&q)
}

func (f *Fq) String() string {
	var s [32]byte
	f.BytesInto(&s)

	// reverse bytes
	for i, j := 0, len(s)-1; i <= j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return hex.EncodeToString(s[:])
}

func montRed(r0, r1, r2, r3, r4, r5, r6, r7 uint64) *Fq {
	k := r0 * INV
	_, carry := futil.Mac(r0, k, q[0], 0)
	r1, carry = futil.Mac(r1, k, q[1], carry)
	r2, carry = futil.Mac(r2, k, q[2], carry)
	r3, carry = futil.Mac(r3, k, q[3], carry)
	r4, carry2 := futil.Adc(r4, 0, carry)

	k = r1 * INV
	_, carry = futil.Mac(r1, k, q[0], 0)
	r2, carry = futil.Mac(r2, k, q[1], carry)
	r3, carry = futil.Mac(r3, k, q[2], carry)
	r4, carry = futil.Mac(r4, k, q[3], carry)
	r5, carry2 = futil.Adc(r5, carry2, carry)

	k = r2 * INV
	_, carry = futil.Mac(r2, k, q[0], 0)
	r3, carry = futil.Mac(r3, k, q[1], carry)
	r4, carry = futil.Mac(r4, k, q[2], carry)
	r5, carry = futil.Mac(r5, k, q[3], carry)
	r6, carry2 = futil.Adc(r6, carry2, carry)

	k = r3 * INV
	_, carry = futil.Mac(r3, k, q[0], 0)
	r4, carry = futil.Mac(r4, k, q[1], carry)
	r5, carry = futil.Mac(r5, k, q[2], carry)
	r6, carry = futil.Mac(r6, k, q[3], carry)
	r7, carry2 = futil.Adc(r7, carry2, carry)

	f := &Fq{r4, r5, r6, r7}

	return f.Sub(&q)

	return f
}