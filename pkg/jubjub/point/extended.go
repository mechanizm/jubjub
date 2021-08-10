package point

import (
	"fmt"

	"github.com/jadeydi/jubjub/pkg/jubjub/fq"
)

type ExtendedPoint struct {
	u, v, z, t1, t2 *fq.Fq
}

func ExtendedFromBytes(byt []byte) *ExtendedPoint {
	affine := AffineFromBytesInner(byt)
	return &ExtendedPoint{
		u:  affine.u,
		v:  affine.v,
		z:  fq.One(),
		t1: affine.u,
		t2: affine.v,
	}
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

func (e *ExtendedPoint) String() string {
	return fmt.Sprintf("u: %s, v: %s, z: %s, t1: %s, t2: %s", e.u.String(), e.v.String(), e.z.String(), e.t1.String(), e.t2.String())
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
