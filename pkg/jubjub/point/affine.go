package point

import (
	"fmt"

	"github.com/jadeydi/jubjub/pkg/jubjub/fq"
)

// AffinePoint represents an affine point `(u, v)` on the
/// curve `-u^2 + v^2 = 1 + d.u^2.v^2` over `Fq` with
/// `d = -(10240/10241)`
type AffinePoint struct {
	u, v fq.Fq
}

// Neg negates the u value in (u,v)
// returning point (-u, v)
func (af *AffinePoint) Neg() *AffinePoint {
	af.u.Neg(af.u)
	return af
}

// from_bytes_inner
func FromBytesInner(byt []byte) (*AffinePoint, error) {
	if len(byt) != 32 {
		return nil, fmt.Errorf("invalid bytes %x", byt)
	}
	sign := byt[31] >> 7
	byt[31] &= 0b0111_1111

	v := fq.FromBytes()
	v2 := v.Square()

	t1 := v2.Sub(fq.One())
	t2 := D.Mul(v2)
}
