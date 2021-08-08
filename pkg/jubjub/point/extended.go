package point

import "github.com/jadeydi/jubjub/pkg/jubjub/fq"

type ExtendedPoint struct {
	u, v, z, t1, t2 fq.Fq
}

func FromBytes(byt []byte) *ExtendedPoint {
	return nil
}
