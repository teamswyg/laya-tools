// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
package compare

import (
	"encoding/hex"
	core "riido.local/next60gjsonsavedcomparison/twoliteral"
	"strconv"
)

func hexValue(v string, max int) bool {
	if len(v) > 2*max || len(v)%2 != 0 {
		return false
	}
	b, e := hex.DecodeString(v)
	return e == nil && hex.EncodeToString(b) == v
}
func validPrimitive(p core.Primitive) bool {
	n, e := strconv.ParseInt(p.Index, 10, 64)
	return p.Type <= 5 && len(p.Raw) <= 12 && e == nil && strconv.FormatInt(n, 10) == p.Index
}
func validateGot(g core.Got) error {
	p := g.Panic
	if !p.Known && p != (core.Panic{}) || p.Known && !p.Present && (p.TypeKnown || p.Type != "" || p.Scope != "") || p.Present && (!p.Known || len(p.Type) > 128 || !p.TypeKnown && p.Type != "" || p.TypeKnown && p.Type == "" || p.Scope != "owned_candidate" && p.Scope != "direct_original_method") {
		return ErrShape
	}
	x := g.Error
	if !x.Available && x != (core.ErrorState{}) || x.Available && !x.Known && (x.Present || x.OwnIdentityKnown || x.OwnIdentity != "" || x.TypeKnown || x.Type != "") {
		return ErrShape
	}
	if x.Available && x.Known {
		if !x.Present {
			if x.OwnIdentityKnown && x.OwnIdentity != "none" || !x.OwnIdentityKnown && x.OwnIdentity != "" || x.TypeKnown || x.Type != "" {
				return ErrShape
			}
		} else {
			if x.TypeKnown && (len(x.Type) < 1 || len(x.Type) > 128) || !x.TypeKnown && x.Type != "" || !x.OwnIdentityKnown && x.OwnIdentity != "" || x.OwnIdentityKnown && x.OwnIdentity != "invalid_line" && x.OwnIdentity != "invalid_utf8" {
				return ErrShape
			}
		}
	}
	if !g.NormalKnown && g.Normal || g.CallbackN < 0 || g.CallbackN > 3 || g.MappedPrefix < 0 || g.MappedPrefix > g.CallbackN || !g.CallbackListCompleteKnown && g.CallbackListComplete || !g.StopKnown && g.StopRequested || !g.FailingLineKnown && g.FailingLine != 0 || g.FailingLine < 0 || g.FailingLine > 4 || !g.TerminalKnown && g.Terminal != "" {
		return ErrShape
	}
	if g.TerminalKnown && g.Terminal != "exhausted" && g.Terminal != "invalid_line" && g.Terminal != "callback_stop" {
		return ErrShape
	}
	for i, c := range g.Callbacks {
		if i >= g.CallbackN {
			if c != (core.Callback{}) {
				return ErrShape
			}
			continue
		}
		if c.PrimitiveKnown && !validPrimitive(c.Primitive) || !c.PrimitiveKnown && c.Primitive != (core.Primitive{}) || !c.LineKnown && c.Line != 0 || c.LineKnown && (c.Line < 1 || c.Line > 4) || !c.ContinueKnown && c.Continue {
			return ErrShape
		}
	}
	if !g.BufferKnown && (g.ReturnedHex != "" || g.ReturnedNil) || g.BufferKnown && !hexValue(g.ReturnedHex, 32) {
		return ErrShape
	}
	if g.BackingBeforeKnown {
		if len(g.BackingBeforeHex) != 64 || !hexValue(g.BackingBeforeHex, 32) {
			return ErrShape
		}
	} else if g.BackingBeforeHex != "" {
		return ErrShape
	}
	if g.BackingAfterKnown {
		if len(g.BackingAfterHex) != 64 || !hexValue(g.BackingAfterHex, 32) {
			return ErrShape
		}
	} else if g.BackingAfterHex != "" {
		return ErrShape
	}
	if g.ActiveInputAfterKnown {
		if !hexValue(g.ActiveInputAfterHex, 1) {
			return ErrShape
		}
	} else if g.ActiveInputAfterHex != "" {
		return ErrShape
	}
	return nil
}
func normalTuple(t [4]int, max int) bool {
	return t[0] >= 0 && t[0] <= max && t[1] == t[0] && t[2] == t[0] && t[3] == 0
}
