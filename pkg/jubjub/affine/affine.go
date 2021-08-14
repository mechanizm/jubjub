package affine

import (
	"fmt"

	"github.com/jadeydi/jubjub/pkg/jubjub/fq"
)

// AffinePoint represents an affine point `(u, v)` on the
/// curve `-u^2 + v^2 = 1 + d.u^2.v^2` over `Fq` with
/// `d = -(10240/10241)`
type AffinePoint struct {
	U, V *fq.Fq
}

// Neg negates the u value in (u,v)
// returning point (-u, v)
func (af *AffinePoint) Neg() *AffinePoint {
	af.U = af.U.Neg()
	return af
}

// from_bytes_inner from_bytes
func FromBytesInner(byt []byte) *AffinePoint {
	sign := byt[31] >> 7
	byt[31] &= 0b0111_1111

	v := fq.FromBytes(byt)
	v2 := v.Square()

	t1 := v2.Sub(fq.One())
	t2 := fq.One().Add(fq.D.Mul(v2))
	u := (t1.Mul(t2.Inverse())).Sqrt()

	flip := (uint64((u.Bytes())[0]) ^ uint64(sign)) & 1
	negated := u.Neg()
	final := fq.ConditionalSelect(u, negated, int(flip))
	return &AffinePoint{
		U: final,
		V: v,
	}
}

func FromRawUnchecked(u, v *fq.Fq) *AffinePoint {
	return &AffinePoint{
		U: u,
		V: v,
	}
}

// IntoBytes converts the af element into its little-endian
// byte representation
func (a *AffinePoint) Bytes() []byte {
	tmp := a.V.Bytes()
	u := a.U.Bytes()

	// Encode the sign of the u-coordinate in the most
	// significant bit.
	tmp[31] |= u[0] << 7

	return tmp[:]
}

func (e *AffinePoint) String() string {
	return fmt.Sprintf("u: %s, v: %s", e.U.String(), e.V.String())
}
