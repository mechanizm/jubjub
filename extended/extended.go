package extended

import (
	"fmt"

	"github.com/mechanizm/jubjub/affine"
	"github.com/mechanizm/jubjub/fq"
)

type ExtendedPoint struct {
	u, v, z, t1, t2 *fq.Fq
}

func FromBytes(byt []byte) *ExtendedPoint {
	a := affine.FromBytesInner(byt)
	return &ExtendedPoint{
		u:  a.U,
		v:  a.V,
		z:  fq.One(),
		t1: a.U,
		t2: a.V,
	}
}

func FromRawUnchecked(u, v *fq.Fq) *ExtendedPoint {
	return FromAffine(affine.FromRawUnchecked(u, v))
}

func (lhs *ExtendedPoint) Add(rhs *ExtendedPoint) *ExtendedPoint {
	return lhs.AddExtendedNiels(rhs.ToNiels())
}

func (e *ExtendedPoint) Mul(buf []byte) *ExtendedPoint {
	return e.ToNiels().Mul(buf)
}

func (e *ExtendedPoint) V() *fq.Fq {
	return e.v
}

// mul_by_cofactor
func (e *ExtendedPoint) MulByCofactor() *ExtendedPoint {
	e = e.Double().Double().Double()
	return e
}

func (e *ExtendedPoint) Double() *ExtendedPoint {
	var uu, vv, zz2, uv2, vvPlusUU, vvMinusUU *fq.Fq

	uu = e.u.Square()
	vv = e.v.Square()
	zz2 = (e.z.Square()).Double()
	uv2 = (e.u.Add(e.v)).Square()

	vvPlusUU = vv.Add(uu)
	vvMinusUU = vv.Sub(uu)

	c := CompletedPoint{
		u: uv2.Sub(vvPlusUU),
		v: vvPlusUU,
		z: vvMinusUU,
		t: zz2.Sub(vvMinusUU),
	}

	return c.Extended()
}

func (e *ExtendedPoint) AddExtendedNiels(other *ExtendedNielsPoint) *ExtendedPoint {
	a := (e.v.Sub(e.u)).Mul(other.VminusU)
	b := (e.v.Add(e.u)).Mul(other.vPlusU)
	c := e.t1.Mul(e.t2).Mul(other.t2d)
	d := (e.z.Mul(other.z)).Double()

	point := &CompletedPoint{
		u: b.Sub(a),
		v: b.Add(a),
		z: d.Add(c),
		t: d.Sub(c),
	}
	return point.Extended()
}

func Identity() *ExtendedPoint {
	return &ExtendedPoint{
		u:  fq.One(),
		v:  fq.One(),
		z:  fq.One(),
		t1: fq.One(),
		t2: fq.One(),
	}
}

func (e *ExtendedPoint) IsIdentity() bool {
	return e.u.Equal(fq.Zero()) && e.v.Equal(e.z)
}

func (e *ExtendedPoint) Bytes() []byte {
	return e.ToAffine().Bytes()
}

func (e *ExtendedPoint) ToAffine() *affine.AffinePoint {
	zinv := e.z.Inverse()
	return &affine.AffinePoint{
		U: e.u.Mul(zinv),
		V: e.v.Mul(zinv),
	}
}

func FromAffine(a *affine.AffinePoint) *ExtendedPoint {
	return &ExtendedPoint{
		u:  a.U,
		v:  a.V,
		z:  fq.One(),
		t1: a.U,
		t2: a.V,
	}
}

type ExtendedNielsPoint struct {
	vPlusU, VminusU, z, t2d *fq.Fq
}

func (e *ExtendedPoint) ToNiels() *ExtendedNielsPoint {
	en := ExtendedNielsPoint{}
	en.VminusU = e.v.Sub(e.u)
	en.vPlusU = e.v.Add(e.u)
	en.z = fq.Set(e.z)
	en.t2d = e.t1.Mul(e.t2).Mul(&fq.EDWARDS_D2)
	return &en
}

func IdentityExtendedNielsPoint() *ExtendedNielsPoint {
	return &ExtendedNielsPoint{
		vPlusU:  fq.One(),
		VminusU: fq.One(),
		z:       fq.One(),
		t2d:     fq.One(),
	}
}

func (niel *ExtendedNielsPoint) Mul(buf []byte) *ExtendedPoint {
	zero := IdentityExtendedNielsPoint()
	acc := Identity()

	var bytes []int
	for i := len(buf) - 1; i >= 0; i-- {
		byt := buf[i]
		for j := 7; j >= 0; j-- {
			bytes = append(bytes, int((byt>>j)&1))
		}
	}

	for _, bit := range bytes[4:] {
		acc = acc.Double()

		acc = acc.AddExtendedNiels(ConditionalSelectExtendedNielsPoint(zero, niel, bit))
	}
	return acc
}

func ConditionalSelectExtendedNielsPoint(a, b *ExtendedNielsPoint, choice int) *ExtendedNielsPoint {
	return &ExtendedNielsPoint{
		vPlusU:  fq.ConditionalSelect(a.vPlusU, b.vPlusU, choice),
		VminusU: fq.ConditionalSelect(a.VminusU, b.VminusU, choice),
		z:       fq.ConditionalSelect(a.z, b.z, choice),
		t2d:     fq.ConditionalSelect(a.t2d, b.t2d, choice),
	}
}

func (e *ExtendedPoint) String() string {
	return fmt.Sprintf("u: %s, v: %s, z: %s, t1: %s, t2: %s", e.u.String(), e.v.String(), e.z.String(), e.t1.String(), e.t2.String())
}

func (e *ExtendedPoint) StringNotCanonical() string {
	return fmt.Sprintf("u: %v, v: %v, z: %v, t1: %v, t2: %v", e.u, e.v, e.z, e.t1, e.t2)
}

type CompletedPoint struct {
	u *fq.Fq
	v *fq.Fq
	z *fq.Fq
	t *fq.Fq
}

func (point *CompletedPoint) Extended() *ExtendedPoint {
	return &ExtendedPoint{
		u:  point.u.Mul(point.t),
		v:  point.v.Mul(point.z),
		z:  point.z.Mul(point.t),
		t1: point.u,
		t2: point.v,
	}
}
