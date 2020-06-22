package crypto

import (
	_ "testing"
)

type printable_parity_data string

type plaintext_ciphertext_pair struct {
	plaintext  printable_parity_data
	ciphertext printable_parity_data
	num_bits   int
}

type testcase struct {
	key              uint64
	uid              uint32
	card_challenge   uint32
	reader_challenge uint32
	reader_response  printable_parity_data
	card_response    printable_parity_data
	data_pairs       int
	data             []plaintext_ciphertext_pair
}

const (
	KEY_1 = 0xffffffffffff
	KEY_2 = 0xa0a1a2a3a4a5
	UID_1 = 0xB479F7D7
	UID_2 = 0x8CBA5DD3
)

var testcases = []testcase{
	/* key, uid, card chal, reader chal,
	reader resp,
	card resp, number of test ciphertexts,
	plaintext/ciphertext pairs ... */
	{KEY_1, UID_1, 0xF3FBAEED, 0x07C9A995,
		"7C 1! 74 1 07 1! EB 1 0F 0! 7B 1 D5 0 1B 0!",
		"3D 1! 0E 1! A0 0! E2 1", 2, []plaintext_ciphertext_pair{
			{"30 1 00 1 02 0 a8 0", "65 0! 8D 0! 65 1 1F 0", 0},
			{"B4 1 79 0 F7 0 D7 1", "52 0 F6 1 46 0 35 1", 0},
		},
	},
	{KEY_1, UID_1, 0x2D4DAAC5, 0x68368F0C,
		"ED 1 73 1! 6B 0 02 1! 88 1 42 1 5B 0 A4 1!",
		"A2 1! D4 0! 3C 0! C3 1", 2, []plaintext_ciphertext_pair{
			{"30 1 00 1 02 0 a8 0", "5B 0 6F 1 96 1 CF 1 ", 0},
			{"B4 1 79 0 F7 0 D7 1", "BB 1 FD 1! 82 1 D2 0!", 0},
		},
	},

	{KEY_1, UID_2, 0x9347B9F4, 0x3BA73C6D,
		"E5 1! 0A 1 5B 0 84 1 44 1 E5 1! C1 0 0C 1",
		"A7 0 A2 1! DA 0 ED 0!", 4, []plaintext_ciphertext_pair{
			{"A0 1 01 0 d6 0 a0 1", "E3 0 B6 0 0E 1! A5 1", 0},
			{"0a", "00", 4},
			{"00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 37 0 49 0",
				"C2 0 E1 1 E4 1 22 0! 99 1 78 0! 6B 0 A1 1! D2 1 C8 1! 62 1! 14 1 0A 1 BA 0 DD 1 AE 0 00 0! 59 0!", 0},
			{"0a", "0C", 4},
		},
	},
	{KEY_2, UID_2, 0x0DF547C9, 0x55414992,
		"85 0 1E 1 29 1! 49 0 BF 0 44 1 5B 1! EB 1",
		"A5 0! 86 1! F4 0 37 1!", 2, []plaintext_ciphertext_pair{
			{"30 1 04 0 26 0 ee 1", "86 1! E0 1! 1B 0! 9E 0", 0},
			{"00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 00 1 37 0 49 0",
				"A3 1 58 1! F2 0 F9 1 00 0! A9 0! 5F 0! A5 1 1C 0 95 0! E7 0! 0D 0 19 0 25 1! F6 0! E1 1 51 0 79 0", 0},
		},
	},
}

func get_hex_value(c rune) int {

	var res = 0

	if c >= 'a' && c <= 'f' {
		res = 10 + int(c) - 'a'
	} else {
		if c >= 'A' && c <= 'F' {
			res = 10 + int(c) - 'A'
		} else {
			res = int(c) - '0'
		}
	}
	return res & 0xf
}
