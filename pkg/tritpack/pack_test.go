// SPDX-License-Identifier: Apache-2.0

package tritpack

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
)

func TestPackedLen(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	cases := []struct {
		count int
		want  int
	}{
		{0, 0}, {1, 1}, {2, 1}, {4, 1}, {5, 1}, {6, 2},
		{18432, 3687},
		{maxInt - 4, (maxInt-4)/5 + boolInt((maxInt-4)%5 != 0)},
		{maxInt - 1, (maxInt-1)/5 + boolInt((maxInt-1)%5 != 0)},
		{maxInt, maxInt/5 + boolInt(maxInt%5 != 0)},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.count), func(t *testing.T) {
			got, err := PackedLen(tc.count)
			if err != nil || got != tc.want {
				t.Fatalf("PackedLen(%d) = (%d, %v), want (%d, nil)", tc.count, got, err, tc.want)
			}
		})
	}
	for _, count := range []int{-1, -5, -maxInt - 1} {
		if got, err := PackedLen(count); err == nil || got != 0 {
			t.Errorf("PackedLen(%d) = (%d, %v), want (0, error)", count, got, err)
		}
	}
}

func TestEveryFiveTritCode(t *testing.T) {
	for code := 0; code < 243; code++ {
		var want [5]int8
		value := code
		for i := range want {
			want[i] = int8(value%3) - 1
			value /= 3
		}
		src := []byte{byte(code)}
		if err := Validate(src, 5); err != nil {
			t.Fatalf("Validate code %d: %v", code, err)
		}
		decoded := []int8{7, 7, 7, 7, 7, 7}
		if n, err := Unpack(decoded, src, 5); err != nil || n != 5 {
			t.Fatalf("Unpack code %d = (%d, %v)", code, n, err)
		}
		if !slices.Equal(decoded[:5], want[:]) || decoded[5] != 7 {
			t.Fatalf("Unpack code %d = %v, want %v with suffix preserved", code, decoded, want)
		}
		encoded := []byte{0xA5, 0xBC}
		if n, err := Pack(encoded, want[:]); err != nil || n != 1 {
			t.Fatalf("Pack code %d = (%d, %v)", code, n, err)
		}
		if encoded[0] != byte(code) || encoded[1] != 0xBC {
			t.Fatalf("Pack code %d = %v, want [%d 188]", code, encoded, code)
		}
		for i, wantTrit := range want {
			if got, err := At(src, i, 5); err != nil || got != wantTrit {
				t.Fatalf("At code %d index %d = (%d, %v), want %d", code, i, got, err, wantTrit)
			}
		}
	}
}

func TestEncodingExamples(t *testing.T) {
	cases := []struct {
		trits []int8
		want  []byte
	}{
		{nil, nil},
		{[]int8{-1}, []byte{120}},
		{[]int8{0}, []byte{121}},
		{[]int8{1}, []byte{122}},
		{[]int8{-1, -1, -1, -1, -1}, []byte{0}},
		{[]int8{1, 1, 1, 1, 1}, []byte{242}},
		{[]int8{1, -1, 0, 1, -1}, []byte{65}},
		{[]int8{-1, -1, -1, -1, -1, 1}, []byte{0, 122}},
	}
	for _, tc := range cases {
		got := make([]byte, len(tc.want))
		if n, err := Pack(got, tc.trits); err != nil || n != len(tc.want) || !bytes.Equal(got, tc.want) {
			t.Errorf("Pack(%v) = (%v, %d, %v), want %v", tc.trits, got, n, err, tc.want)
		}
	}
}

func TestRoundTripCounts(t *testing.T) {
	for _, count := range []int{0, 1, 2, 4, 5, 6, 18432} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			src := make([]int8, count)
			for i := range src {
				src[i] = int8((i+i/7)%3) - 1
			}
			original := slices.Clone(src)
			length, err := PackedLen(count)
			if err != nil {
				t.Fatal(err)
			}
			encoded := bytes.Repeat([]byte{0xD3}, length+3)
			if n, err := Pack(encoded, src); err != nil || n != length {
				t.Fatalf("Pack = (%d, %v), want (%d, nil)", n, err, length)
			}
			if !bytes.Equal(encoded[length:], []byte{0xD3, 0xD3, 0xD3}) {
				t.Fatalf("Pack changed suffix: %v", encoded[length:])
			}
			if !slices.Equal(src, original) {
				t.Fatal("Pack changed source trits")
			}
			encoded = encoded[:length]
			if err := Validate(encoded, count); err != nil {
				t.Fatalf("Validate packed input: %v", err)
			}
			decoded := make([]int8, count+3)
			for i := range decoded {
				decoded[i] = 7
			}
			if n, err := Unpack(decoded, encoded, count); err != nil || n != count {
				t.Fatalf("Unpack = (%d, %v), want (%d, nil)", n, err, count)
			}
			if !slices.Equal(decoded[:count], src) || !slices.Equal(decoded[count:], []int8{7, 7, 7}) {
				t.Fatal("Unpack did not round trip while preserving suffix")
			}
			for i, want := range src {
				if got, err := At(encoded, i, count); err != nil || got != want {
					t.Fatalf("At(%d) = (%d, %v), want %d", i, got, err, want)
				}
			}
		})
	}
}

func TestCanonicalPadding(t *testing.T) {
	// A code is canonical for a partial group precisely when every omitted
	// base-3 digit is 1. Check all valid byte values for each partial length.
	for count := 1; count < 5; count++ {
		for code := 0; code < 243; code++ {
			value := code
			for i := 0; i < count; i++ {
				value /= 3
			}
			canonical := true
			for i := count; i < 5; i++ {
				canonical = canonical && value%3 == 1
				value /= 3
			}
			src := []byte{byte(code)}
			if err := Validate(src, count); (err == nil) != canonical {
				t.Fatalf("Validate code %d count %d = %v, canonical=%v", code, count, err, canonical)
			}
			for index := 0; index < count; index++ {
				if _, err := At(src, index, count); (err == nil) != canonical {
					t.Fatalf("At code %d count %d index %d = %v, canonical=%v", code, count, index, err, canonical)
				}
			}
		}
	}
}

func TestRejectInvalidBytes(t *testing.T) {
	for code := 243; code <= 255; code++ {
		for _, byteIndex := range []int{0, 1} {
			src := []byte{121, 121}
			src[byteIndex] = byte(code)
			before := bytes.Clone(src)
			if err := Validate(src, 10); err == nil {
				t.Fatalf("Validate accepted invalid byte %d at %d", code, byteIndex)
			}
			dst := []int8{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}
			original := slices.Clone(dst)
			if n, err := Unpack(dst, src, 10); err == nil || n != 0 || !slices.Equal(dst, original) {
				t.Fatalf("Unpack invalid byte %d at %d = (%d, %v), dst=%v", code, byteIndex, n, err, dst)
			}
			if got, err := At(src, byteIndex*5, 10); err == nil || got != 0 {
				t.Fatalf("At invalid byte %d at %d = (%d, %v)", code, byteIndex, got, err)
			}
			if !bytes.Equal(src, before) {
				t.Fatal("validation or decoding changed encoded input")
			}
		}
	}
}

func TestPackErrorsPreserveDestination(t *testing.T) {
	for trit := -128; trit <= 127; trit++ {
		if trit >= -1 && trit <= 1 {
			continue
		}
		// The invalid value occurs after a complete byte to catch partial writes.
		src := []int8{0, 0, 0, 0, 0, 0, int8(trit)}
		original := slices.Clone(src)
		dst := []byte{0xA5, 0xA5, 0xA5}
		before := bytes.Clone(dst)
		if n, err := Pack(dst, src); err == nil || n != 0 || !bytes.Equal(dst, before) {
			t.Fatalf("Pack invalid trit %d = (%d, %v), dst=%v", trit, n, err, dst)
		}
		if !slices.Equal(src, original) {
			t.Fatal("Pack changed source on error")
		}
	}
	for _, dst := range [][]byte{nil, {0xA5}} {
		before := bytes.Clone(dst)
		if n, err := Pack(dst, []int8{0, 0, 0, 0, 0, 0}); err == nil || n != 0 || !bytes.Equal(dst, before) {
			t.Fatalf("Pack short destination = (%d, %v), dst=%v", n, err, dst)
		}
	}
}

func TestUnpackErrorsPreserveDestination(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	cases := []struct {
		name  string
		src   []byte
		count int
		size  int
	}{
		{"negative count", nil, -1, 3},
		{"minimum count", nil, -maxInt - 1, 3},
		{"maximum count", nil, maxInt, 3},
		{"missing byte", []byte{121}, 6, 9},
		{"trailing byte", []byte{121, 121}, 5, 9},
		{"byte for empty", []byte{121}, 0, 3},
		{"padding after complete group", []byte{121, 0}, 6, 9},
		{"positive padding", []byte{202}, 4, 9},
		{"negative padding", []byte{40}, 4, 9},
		{"short destination", []byte{121, 121}, 6, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dst := make([]int8, tc.size)
			for i := range dst {
				dst[i] = 7
			}
			before := slices.Clone(dst)
			original := bytes.Clone(tc.src)
			if n, err := Unpack(dst, tc.src, tc.count); err == nil || n != 0 {
				t.Fatalf("Unpack = (%d, %v), want (0, error)", n, err)
			}
			if !slices.Equal(dst, before) || !bytes.Equal(tc.src, original) {
				t.Fatal("Unpack changed a buffer on error")
			}
		})
	}
}

func TestValidateLengthAndCount(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	cases := []struct {
		src   []byte
		count int
	}{
		{nil, -1}, {nil, -maxInt - 1}, {nil, maxInt},
		{nil, 1}, {[]byte{121}, 6}, {[]byte{121, 121}, 5},
		{[]byte{121}, 0}, {[]byte{121, 121, 121}, 6},
	}
	for _, tc := range cases {
		if err := Validate(tc.src, tc.count); err == nil {
			t.Errorf("Validate(%v, %d) succeeded", tc.src, tc.count)
		}
	}
	if err := Validate(nil, 0); err != nil {
		t.Fatalf("Validate empty input: %v", err)
	}
}

func TestAtBoundsAndLength(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	cases := []struct {
		src          []byte
		index, count int
	}{
		{nil, 0, 0}, {[]byte{121}, -1, 5}, {[]byte{121}, -maxInt - 1, 5},
		{[]byte{121}, 5, 5}, {[]byte{121}, maxInt, 5},
		{nil, 0, -1}, {nil, 0, -maxInt - 1}, {nil, 0, maxInt},
		{nil, 0, 1}, {[]byte{121}, 0, 6}, {[]byte{121, 121}, 0, 5},
	}
	for _, tc := range cases {
		if got, err := At(tc.src, tc.index, tc.count); err == nil || got != 0 {
			t.Errorf("At(%v, %d, %d) = (%d, %v), want (0, error)", tc.src, tc.index, tc.count, got, err)
		}
	}
}

func TestAtChecksOnlyAddressedByte(t *testing.T) {
	cases := []struct {
		src          []byte
		index, count int
	}{
		{[]byte{243, 121}, 5, 10},
		{[]byte{121, 255}, 0, 10},
		{[]byte{121, 0}, 0, 6}, // Noncanonical padding is in another byte.
	}
	for _, tc := range cases {
		if err := Validate(tc.src, tc.count); err == nil {
			t.Fatalf("test input should fail full validation: %v", tc.src)
		}
		if got, err := At(tc.src, tc.index, tc.count); err != nil || got != 0 {
			t.Errorf("At(%v, %d, %d) = (%d, %v), want (0, nil)", tc.src, tc.index, tc.count, got, err)
		}
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
