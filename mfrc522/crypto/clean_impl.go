package crypto

/**
  Clean crypto1 implementation
*/

/* == API section =============================== */
/* Initialize the LFSR with the key */
func crypto1_clean_init(state *Crypto1State, key uint64) {
	state.lfsr = 0
	state.prng = 0
	k := key
	for i := 0; i < 6; i++ {
		state.lfsr <<= 8
		state.lfsr |= k & 0xff
		k >>= 8
	}
}

/* Shift UID xor card_nonce into the LFSR without active cipher stream
 * feedback */
func crypto1_clean_mutual_1(state *Crypto1State, uid, card_challenge uint32) {
	IV := uid ^ card_challenge
	/* Go through the IV bytes in MSByte first, LSBit first order */
	mifare_update_word(state, IV, 0)
	state.prng = uint16(card_challenge) /* Load the card’s PRNG state into our PRNG */
}

/* Shift in the reader nonce to generate the reader challenge,
* then generate the reader response */
func crypto1_clean_mutual_2_reader(state *Crypto1State, reader_response []parity_data) bool {
	/* Unencrypted reader nonce */
	reader_nonce := array_to_uint32(reader_response)

	/* Feed the reader nonce into the state and simultaneously encrypt it */
	for i := 3; i >= 0; i-- { /* Same as in mifare_update_word, but with added parity */
		reader_response[3-i] = reader_response[3-i] ^
			parity_data(mifare_update_byte(state, uint8((reader_nonce>>(i*8))&0xff), 0))
		reader_response[3-i] ^= parity_data(mf20(state.lfsr) << 8)
	}

	/* Unencrypted reader response */
	rr := prng_next(state, 64)

	uint32_to_array_with_parity(rr, reader_response[4:])

	/* Encrypt the reader response */
	state.Transcrypt(reader_response[4:], 4)

	return true
}

/* Generate the expected card response and compare it to the actual
* card response */
func crypto1_clean_mutual_3_reader(state *Crypto1State, card_response []parity_data, bytes int) bool {
	TR_is := array_to_uint32(card_response)
	TR_should := prng_next(state, 96) ^ mifare_update_word(state, 0, 0)
	return TR_is == TR_should
}

/* Shift in the reader challenge into the state, generate expected reader
* response and compare it to actual reader response. */
func crypto1_clean_mutual_2_card(state *Crypto1State, reader_response []parity_data) bool {

	/* Reader challenge/Encrypted reader nonce */
	RC := array_to_uint32(reader_response)

	/* Encrypted reader response */
	RR_is := array_to_uint32(reader_response[4:])

	/* Shift in reader challenge */
	// keystream := mifare_update_word(state, RC, 1)
	// correct reader challenge: RC ^ keystream
	mifare_update_word(state, RC, 1)

	/* Generate expected reader response */
	RR_should := prng_next(state, 64) ^ mifare_update_word(state, 0, 0)
	return RR_should == RR_is
}

/* Output the card response */
func crypto1_clean_mutual_3_card(state *Crypto1State, card_response []parity_data, bytes int) bool {
	/* Unencrypted tag response */
	TR := prng_next(state, 96)
	uint32_to_array_with_parity(TR, card_response)

	/* Encrypt the response */
	state.Transcrypt(card_response, bytes)
	return true
}

/* Encrypt or decrypt a number of bytes */
func crypto1_clean_transcrypt_bits(state *Crypto1State, data []parity_data, bytes, bits int) {
	for i := 0; i < bytes; i++ {
		data[i] = data[i] ^ parity_data(mifare_update_byte(state, 0, 0))
		data[i] = data[i] ^ parity_data((mf20(state.lfsr) << 8))
	}

	for i := 0; i < bits; i++ {
		data[bytes] ^= parity_data(mifare_update(state, 0, 0) << i)
	}
}

func CleanCrypto1ReaderInit() Crypto1State {
	return Crypto1State{
		init:     crypto1_clean_init,
		mutual_1: crypto1_clean_mutual_1,
		mutual_2: crypto1_clean_mutual_2_reader,
		mutual_3: crypto1_clean_mutual_3_reader,
	}
}

func CleanCrypto1CardInit() Crypto1State {
	return Crypto1State{
		init:     crypto1_clean_init,
		mutual_1: crypto1_clean_mutual_1,
		mutual_2: crypto1_clean_mutual_2_card,
		mutual_3: crypto1_clean_mutual_3_card,
	}
}
