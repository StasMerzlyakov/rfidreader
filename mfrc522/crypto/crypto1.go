package crypto

import (
	"math/bits"
)

/**
 * Mifare Classic – Eine Analyse der
 * Implementierung
 * https://sar.informatik.hu-berlin.de/research/publications/SAR-PR-2008-21/SAR-PR-2008-21_.pdf
 */

const (
	mf2_f4a = 0x9E98
	mf2_f4b = 0xB48E
	mf2_f5c = 0xEC57E80A
)

/*
 * This type holds data bytes with associated parity bits.
 * The data is in the low byte while the associated parity bit
 * is in the least-significant bit of the high byte.
 */
type parity_data uint16

type Crypto1State struct {
	/* The 48 bit LFSR for the main cipher state and keystream generation */
	lfsr uint64

	/* The 16 bit LFSR for the card PRNG state, also used during authentication.*/
	prng uint16

	/* Initialize a cipher instance with secret key */
	init func(state *Crypto1State, key uint64)

	/*
	 * First stage of mutual authentication given a card’s UID.
	 * card_challenge is the card nonce as an integer
	 */
	mutual_1 func(state *Crypto1State, uid, card_challenge uint32)

	/*
	* Second stage of mutual authentication.
	* If this is the reader side, then the first 4 bytes of reader_response must
	* be preloaded with the reader nonce (and parity) and all 8 bytes will be
	* computed to be the correct reader response to the card challenge.
	* If this is the card side, then the response to the card challenge will be
	* checked.
	 */
	mutual_2 func(state *Crypto1State, reader_response []parity_data) bool

	/*
	* Third stage of mutual authentication.
	* If this is the reader side, then the card response to the reader
	* challenge will be checked.
	* If this is the card side, then the card response to the reader
	* challenge will be computed.
	 */
	mutual_3 func(state *Crypto1State, card_response []parity_data, bytes int) bool

	/**
	* Perform the Crypto-1 encryption or decryption operation on ’bytes’ bytes
	* of data with associated parity bits.
	* The additional parameter ’bits’ allows processing incomplete bytes after the
	* last byte. That is, if bits > 0 then data should contain (bytes+1) bytes where
	* the last byte is incomplete.
	 */
	transcrypt_bits func(state *Crypto1State, card_response []parity_data, bytes, bits int)
}

func (state *Crypto1State) Init(key uint64) {
	state.init(state, key)
}

func (state *Crypto1State) Mutual_1(uid, card_challenge uint32) {
	state.mutual_1(state, uid, card_challenge)
}

func (state *Crypto1State) Mutual_2(reader_response []parity_data) {
	state.mutual_2(state, reader_response)
}

func (state *Crypto1State) Mutual_3(card_response []parity_data, bytes int) {
	state.mutual_3(state, card_response, bytes)
}

func (state *Crypto1State) Transcrypt_bits(card_response []parity_data, bytes, bits int) {
	state.transcrypt_bits(state, card_response, bytes, bits)
}

func (state *Crypto1State) Transcrypt(card_response []parity_data, bytes int) {
	state.transcrypt_bits(state, card_response, bytes, 0)
}

/* == PRNG function ========================== */
/* Clock the prng register by n steps and return the new state, don’t
 * update the register.
 * Note: returns a 32 bit value, even if the register is only 16 bit wide.
 * This return value is only valid, when the register was clocked at least
 * 16 times. */
func prng_next(state *Crypto1State, n int) uint32 {
	prng := uint32(state.prng)

	/* The register is stored and returned in reverse bit order, this way, even
	 * if we cast the returned 32 bit value to a 16 bit value, the necessary
	 * state will be retained. */
	prng = bits.Reverse32(prng)

	for i := 0; i < n; i++ {
		prng = (prng << 1) | (((prng >> 15) ^ (prng >> 13) ^ (prng >> 12) ^ (prng >> 10)) & 1)
	}
	return bits.Reverse32(prng)

}

/* == keystream generating filter function === */
/* This function selects the four bits at offset a, b, c and d from the value x
 * and returns the concatenated bitstring x_d || x_c || x_b || x_a as an integer
 */
func i4(x uint64, a, b, c, d int) uint32 {
	return uint32((x >> a & 1) | (x>>b&1)<<1 | (x>>c&1)<<2 | (x>>c&1)<<3)
}

/* Return the nth bit from x */
func bit(x, n uint8) uint8 {
	return (x >> n) & 1
}

/* Convert 4 array entries (a[0], a[1], a[2] and a[3]) into a 32 bit integer,
* where a[0] is the MSByte and a[3] is the LSByte */
func array_to_uint32(a []parity_data) uint32 {
	return uint32(a[0]&0xff)<<24 |
		uint32(a[1]&0xff)<<16 |
		uint32(a[2]&0xff)<<8 |
		uint32(a[3]&0xff)<<0
}

/* Calculate the odd parity bit for one byte of input */
func odd_parity(i parity_data) parity_data {
	return ((i ^ i>>1 ^ i>>2 ^ i>>3 ^ i>>4 ^ i>>5 ^ i>>6 ^ i>>7 ^ 1) & 0x01)
}

func uint32_to_array_with_parity(i uint32, a []parity_data) {
	a[0] = parity_data((i >> 24) & 0xff)
	a[0] |= odd_parity(a[0]) << 8

	a[1] = parity_data((i >> 16) & 0xff)
	a[1] |= odd_parity(a[1]) << 8

	a[2] = parity_data((i >> 8) & 0xff)
	a[2] |= odd_parity(a[2]) << 8

	a[3] = parity_data((i >> 0) & 0xff)
	a[3] |= odd_parity(a[0]) << 8
}

//UINT32_TO_ARRAY_WITH_PARITY

/* Return one bit of non-linear filter function output for 48 bits of
 * state input */
func mf20(x uint64) uint32 {
	/* number of cycles between when key stream is produced
	* and when key stream is used.
	* Irrelevant for software implementations, but important
	* to consider in side-channel attacks */
	d := 2

	i5 := uint32(((mf2_f4b>>i4(x, 7+d, 9+d, 11+d, 13+d))&1)<<0 |
		((mf2_f4a>>i4(x, 15+d, 17+d, 19+d, 21+d))&1)<<1 |
		((mf2_f4a>>i4(x, 23+d, 25+d, 27+d, 29+d))&1)<<2 |
		((mf2_f4b>>i4(x, 31+d, 33+d, 35+d, 37+d))&1)<<3 |
		((mf2_f4a>>i4(x, 39+d, 41+d, 43+d, 45+d))&1)<<4)
	return (mf2_f5c >> i5) & 1
}

/* == LFSR state update functions ============ */
/* Updates the 48-bit LFSR in state using the mifare taps, optionally
 * XORing in 1 bit of additional input, optionally XORing in 1 bit of
 * cipher stream output (e.g. feeding back the output).
 * Return current cipher stream output bit. */
func mifare_update(state *Crypto1State, injection, feedback uint8) uint8 {
	x := state.lfsr
	ks := mf20(state.lfsr)
	fb := uint32(0)
	if feedback > 0 {
		fb = ks
	}
	state.lfsr = (x >> 1) |
		((((x >> 0) ^ (x >> 5) ^
			(x >> 9) ^ (x >> 10) ^ (x >> 12) ^ (x >> 14) ^
			(x >> 15) ^ (x >> 17) ^ (x >> 19) ^ (x >> 24) ^
			(x >> 25) ^ (x >> 27) ^ (x >> 29) ^ (x >> 35) ^
			(x >> 39) ^ (x >> 41) ^ (x >> 42) ^ (x >> 43) ^
			uint64(uint32(injection)^fb)) & 1) << 47)
	return uint8(ks)
}

/* Update the 48-bit LFSR in state using the mifare taps 8 times, optionally
 * XORing in 1 bit of additional input per step (LSBit first).
 * Return corresponding cipher stream. */
func mifare_update_byte(state *Crypto1State, injection, feedback uint8) uint8 {
	ret := uint8(0)
	for i := uint8(0); i < 8; i++ {
		ret |= mifare_update(state, bit(injection, i), feedback) << i
	}
	return ret
}

/* Update the 48-bit LFSR in state using the mifare taps 32 times, optionally
 * XORing in 1 byte of additional input per step (MSByte first).
 * Return the corresponding cipher stream. */
func mifare_update_word(state *Crypto1State, injection uint32, feedback uint8) uint32 {
	ret := uint32(0)
	for i := 3; i >= 0; i-- {
		ret |= uint32(mifare_update_byte(state, uint8((injection>>(i*8))&0xff), feedback) << (i * 8))
	}
	return ret
}
