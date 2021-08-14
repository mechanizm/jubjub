package fr

// INV = -(r^{-1} mod 2^64) mod 2^64
const INV uint64 = 0x1ba3a358ef788ef9

const MODULUS_BITS uint32 = 252

const NUM_BITS uint32 = MODULUS_BITS

// r is modulus in Fr
// r = 0x0e7db4ea6533afa906673b0101343b00a6682093ccc81082d0970e5ed6f72cb7
var r = Fr{
	0xd0970e5ed6f72cb7,
	0xa6682093ccc81082,
	0x06673b0101343b00,
	0x0e7db4ea6533afa9,
}

/// R = 2^256 mod r
var R = Fr{
	0x25f8_0bb3_b996_07d9,
	0xf315_d62f_66b6_e750,
	0x9325_14ee_eb88_14f4,
	0x09a6_fc6f_4791_55c6,
}

/// R^2 = 2^512 mod r
var R2 = Fr{
	0x67719aa495e57731,
	0x51b0cef09ce3fc26,
	0x69dab7fac026e9a5,
	0x04f6547b8d127688,
}

/// R^3 = 2^768 mod r
var R3 = Fr{
	0xe0d6c6563d830544,
	0x323e3883598d0f85,
	0xf0fea3004c2e2ba8,
	0x05874f84946737ec,
}

var MODULUS = Fr{
	0xd097_0e5e_d6f7_2cb7,
	0xa668_2093_ccc8_1082,
	0x0667_3b01_0134_3b00,
	0x0e7d_b4ea_6533_afa9,
}
