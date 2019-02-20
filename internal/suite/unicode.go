package suite

import (
	"fmt"
	"unicode"

	. "github.com/dmolesUC3/cos/pkg"

	"github.com/dmolesUC3/cos/internal/objects"
)

const (
	keyMaxBytes      = 1024
	maxRunesToReport = 40
)

func AllUnicodeCases() []Case {
	var cases []Case
	cases = append(cases, UnicodePropertiesCases()...)
	cases = append(cases, UnicodeScriptsCases()...)
	return cases
}

func UnicodePropertiesCases() []Case {
	return toCases("Unicode properties: ", unicode.Properties)
}

func UnicodeScriptsCases() []Case {
	return toCases("Unicode scripts: ", unicode.Scripts)
}

func toCases(prefix string, tables map[string]*unicode.RangeTable) []Case {
	var cases []Case
	for rangeName, rt := range tables {
		// Bad things happen if we try to cast these to runes
		if rt == unicode.Noncharacter_Code_Point {
			continue
		}
		uc := newUnicodeCase(prefix+rangeName, rangeTableToRunes(rt))
		cases = append(cases, uc)
	}
	return cases
}

type unicodeCase struct {
	caseImpl
	allRunes []rune
}

func newUnicodeCase(rangeName string, allRunes []rune) Case {
	c := unicodeCase{allRunes: allRunes}
	c.name = fmt.Sprintf("%v (%d characters)", rangeName, len(allRunes))
	c.exec = c.doExec
	return &c
}

func (u *unicodeCase) doExec(target objects.Target) (ok bool, detail string) {
	invalidRunesForKey := findInvalidRunesForKeyIn(u.allRunes, target)
	numInvalidRunes := len(invalidRunesForKey)
	if numInvalidRunes == 0 {
		return true, "no invalid characters found"
	}
	var invalidRunesStr string
	if numInvalidRunes < maxRunesToReport {
		invalidRunesStr = string(invalidRunesForKey)
	} else {
		invalidRunesStr = string(invalidRunesForKey[0:maxRunesToReport]) + "…"
	}
	return true, fmt.Sprintf("%d invalid characters: %#v", numInvalidRunes, invalidRunesStr)
}

// TODO: parallelize this?
func findInvalidRunesForKeyIn(keyRunes []rune, target objects.Target) []rune {
	if len(keyRunes) == 0 {
		return nil
	}
	if len(keyRunes) < keyMaxBytes {
		crvd := NewCrvd(target, string(keyRunes), DefaultContentLengthBytes, DefaultRandomSeed)
		err := crvd.CreateRetrieveVerifyDelete()
		if err == nil {
			return nil
		}
		runes := []rune(keyRunes)
		if len(runes) == 1 {
			return runes
		}
	}
	// Either:
	// 1. we have too many characters to test in a single key, so we split it, or
	// 2. we have one or more invalid key characters somewhere in this string, so we binary search for them
	kr1, kr2 := split(keyRunes)
	result1 := findInvalidRunesForKeyIn(kr1, target)
	result2 := findInvalidRunesForKeyIn(kr2, target)
	return append(result1, result2...)
}

func split(s []rune) (left, right []rune) {
	r := []rune(s)
	left = r[0 : len(r)/2]
	right = r[len(r[0:len(r)/2]):]
	return left, right
}

func rangeTableToRunes(rt *unicode.RangeTable) []rune {
	var runes []rune
	for _, r16 := range rt.R16 {
		runes = append(runes, range16ToRunes(r16)...)
	}
	for _, r32 := range rt.R32 {
		runes = append(runes, range32ToRunes(r32)...)
	}
	return runes
}

func range16ToRunes(r16 unicode.Range16) []rune {
	lenRunes16 := (1 + r16.Hi - r16.Lo) / r16.Stride
	var runes = make([]rune, lenRunes16)
	start16 := 0
	for i, cp := start16, r16.Lo; cp <= r16.Hi; cp += r16.Stride {
		runes[i] = rune(cp)
	}
	return runes
}

func range32ToRunes(r32 unicode.Range32) []rune {
	lenRunes32 := (1 + r32.Hi - r32.Lo) / r32.Stride
	var runes = make([]rune, lenRunes32)
	for i, cp := 0, r32.Lo; cp <= r32.Hi; cp += r32.Stride {
		runes[i] = rune(cp)
	}
	return runes
}