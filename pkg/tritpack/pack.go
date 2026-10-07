// SPDX-License-Identifier: Apache-2.0

// Package tritpack packs signed trits (-1, 0, +1) into bytes.
//
// Each byte holds five trits as little-endian base-3 digits: digit = trit + 1,
// so the first trit has weight 1, then 3, 9, 27, and 81. The valid byte values
// are 0 through 242. Unused digits in the final byte must be 1 (the zero trit),
// and an encoding has exactly ceil(count/5) bytes.
//
// Complete groups occupy 1.6 physical bits per trit (8/5), distinct from the
// information limit log2(3), approximately 1.585 bits per trit. Partial groups
// require a whole byte with canonical padding.
package tritpack

import "fmt"

const (
	tritsPerByte = 5
	maxCode      = 242
	zeroCode     = 121 // 1 + 3 + 9 + 27 + 81: five zero trits.
)

// PackedLen returns the exact encoded length for n trits. It rejects negative
// counts and handles every nonnegative int without overflowing.
func PackedLen(n int) (int, error) {
	if n < 0 {
		return 0, fmt.Errorf("tritpack: negative trit count %d", n)
	}
	length := n / tritsPerByte
	if n%tritsPerByte != 0 {
		length++
	}
	return length, nil
}

// Pack encodes src into dst and returns the number of bytes written. dst must
// have at least PackedLen(len(src)) bytes. It validates all trits before writing,
// returns zero on error, and preserves bytes after the encoded length.
func Pack(dst []byte, src []int8) (int, error) {
	length, err := PackedLen(len(src))
	if err != nil {
		return 0, err
	}
	if len(dst) < length {
		return 0, fmt.Errorf("tritpack: destination has %d bytes, need %d", len(dst), length)
	}
	for i, trit := range src {
		if trit < -1 || trit > 1 {
			return 0, fmt.Errorf("tritpack: invalid trit %d at index %d", trit, i)
		}
	}
	for byteIndex, tritIndex := 0, 0; byteIndex < length; byteIndex++ {
		code := zeroCode
		for digit, weight := 0, 1; digit < tritsPerByte && tritIndex < len(src); digit, weight = digit+1, weight*3 {
			code += int(src[tritIndex]) * weight
			tritIndex++
		}
		dst[byteIndex] = byte(code)
	}
	return length, nil
}

// Unpack decodes exactly count trits from src into dst and returns the number
// written. src must be a canonical encoding of count trits, and dst must have
// at least count entries. It validates the whole input before writing, returns
// zero on error, and preserves entries after count.
func Unpack(dst []int8, src []byte, count int) (int, error) {
	if err := Validate(src, count); err != nil {
		return 0, err
	}
	if len(dst) < count {
		return 0, fmt.Errorf("tritpack: destination has %d trits, need %d", len(dst), count)
	}
	for byteIndex, tritIndex := 0, 0; tritIndex < count; byteIndex++ {
		code := src[byteIndex]
		for digit := 0; digit < tritsPerByte && tritIndex < count; digit++ {
			dst[tritIndex] = int8(code%3) - 1
			code /= 3
			tritIndex++
		}
	}
	return count, nil
}

// Validate checks the exact encoded length, every byte, and canonical padding
// for count trits. It does not modify src.
func Validate(src []byte, count int) error {
	if err := validateLength(src, count); err != nil {
		return err
	}
	for i, code := range src {
		if code > maxCode {
			return fmt.Errorf("tritpack: invalid byte %d at index %d", code, i)
		}
	}
	if count%tritsPerByte != 0 {
		return validatePadding(src[len(src)-1], count%tritsPerByte)
	}
	return nil
}

// At returns the trit at index in an encoding of count trits. It checks the
// count, exact encoded length, index, and addressed byte, including padding if
// that byte is the final partial group. It does not inspect any other byte;
// callers should Validate an entire encoding once when loading it.
func At(src []byte, index, count int) (int8, error) {
	if err := validateLength(src, count); err != nil {
		return 0, err
	}
	if index < 0 || index >= count {
		return 0, fmt.Errorf("tritpack: index %d outside trit count %d", index, count)
	}
	byteIndex := index / tritsPerByte
	code := src[byteIndex]
	if code > maxCode {
		return 0, fmt.Errorf("tritpack: invalid byte %d at index %d", code, byteIndex)
	}
	if byteIndex == len(src)-1 && count%tritsPerByte != 0 {
		if err := validatePadding(code, count%tritsPerByte); err != nil {
			return 0, err
		}
	}
	for digit := 0; digit < index%tritsPerByte; digit++ {
		code /= 3
	}
	return int8(code%3) - 1, nil
}

func validateLength(src []byte, count int) error {
	length, err := PackedLen(count)
	if err != nil {
		return err
	}
	if len(src) != length {
		return fmt.Errorf("tritpack: encoded length %d, need exactly %d for %d trits", len(src), length, count)
	}
	return nil
}

func validatePadding(code byte, used int) error {
	for digit := 0; digit < used; digit++ {
		code /= 3
	}
	for digit := used; digit < tritsPerByte; digit++ {
		if code%3 != 1 {
			return fmt.Errorf("tritpack: noncanonical padding digit %d", digit)
		}
		code /= 3
	}
	return nil
}
