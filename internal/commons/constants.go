package commons

import hpke "github.com/danielhavir/go-hpke"

// HpkeMode specifies ciphersuite
const HpkeMode = hpke.BASE_X25519_SHA256_ChaCha20Poly1305

// BufferSize specifies buffer during file exchange
const BufferSize = 4096
