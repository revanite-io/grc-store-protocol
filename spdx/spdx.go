// SPDX-License-Identifier: Apache-2.0

// Package spdx validates and canonicalizes SPDX license expressions for the
// grc.store publication-license field (ADR-0036).
//
// (The package lives at spdx/ rather than license/ because the module's
// case-insensitive checkout cannot hold both a license/ directory and the
// top-level LICENSE file.)
//
// A publisher declares a publication license at publish time; grcli stamps it
// as the OCI manifest annotation org.opencontainers.image.licenses and the hub
// reads it back. Both ends must agree on what a valid license is and on its
// canonical spelling, or the index fragments (Apache-2.0 vs apache-2.0 vs ASL2
// as three "licenses"). This package is that one shared definition.
//
// The two ends use it differently, by design (ADR-0036 decision 4):
//
//   - grcli is the STRICT gate. It calls Canonicalize and hard-fails the publish
//     before any network call on a malformed expression OR an unknown leaf id —
//     so typos never reach the registry.
//   - The hub is LENIENT. It calls Parse (syntax only) and stores
//     Expression.String() — the structurally-canonicalized form, with known ids
//     re-cased but unknown-but-well-formed ids passed through unchanged. It never
//     rejects a sync over the license field: a publisher on a newer SPDX list than
//     the hub must not have their already-pushed bytes orphaned. A malformed
//     annotation parses with an error and the hub stores NULL.
//
// Canonicalization (String) is what protects the index, not rejection: it
// collapses divergent spellings of the SAME license regardless of which path ran.
//
// SPDX grammar supported: simple ids, the "or-later" "+" suffix, "WITH" exception,
// "AND"/"OR" (OR binds loosest), parentheses, and LicenseRef-/DocumentRef- custom
// identifiers. Operators are case-sensitive uppercase per the SPDX spec.
//
// STABILITY: the canonical OUTPUT of String for a given input is contract — two
// importers on the same SPDXListVersion must produce byte-identical output.
// Bumping SPDXListVersion is additive (it can only make a previously-unknown id
// known, never change a known id's canonical casing), so it is a minor release;
// any change that alters String's output for an already-known id is BREAKING.
//
//go:generate go run spdxgen.go
package spdx

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrSyntax means the string is not a well-formed SPDX license expression
// (the grammar is violated). The hub treats this as "no license" (store NULL).
var ErrSyntax = errors.New("malformed SPDX license expression")

// ErrUnknownID means the expression is well-formed but contains a license or
// exception id not in SPDXListVersion's list (excluding LicenseRef-/DocumentRef-,
// which are custom by definition). grcli treats this as fatal; the hub does not.
var ErrUnknownID = errors.New("unknown SPDX license or exception id")

// Canonicalize is the strict entry point (grcli's gate). It returns the canonical
// form of a well-formed expression whose every leaf id is known to
// SPDXListVersion, or an error wrapping ErrSyntax / ErrUnknownID otherwise.
func Canonicalize(expr string) (string, error) {
	e, err := Parse(expr)
	if err != nil {
		return "", err
	}
	if unknown := e.UnknownIDs(); len(unknown) > 0 {
		return "", fmt.Errorf("%w: %s", ErrUnknownID, strings.Join(unknown, ", "))
	}
	return e.String(), nil
}

// Parse validates SPDX expression GRAMMAR only and returns the parsed expression.
// It does NOT fail on unrecognized (but well-formed) ids — call UnknownIDs to
// check those. This is the hub's lenient entry point: on success store String();
// on error store NULL.
func Parse(expr string) (*Expression, error) {
	p := &parser{toks: tokenize(expr)}
	node, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if !p.atEnd() {
		return nil, fmt.Errorf("%w: unexpected %q", ErrSyntax, p.peek())
	}
	return &Expression{root: node}, nil
}

// Expression is a parsed SPDX license expression. Its zero value is not usable;
// obtain one from Parse.
type Expression struct{ root node }

// String returns the canonical form: operators uppercased and single-spaced,
// known ids re-cased to their official SPDX spelling, "+" preserved,
// LicenseRef-/DocumentRef- preserved verbatim (those are case-sensitive), and
// the minimum parentheses needed to preserve OR/AND precedence.
func (e *Expression) String() string {
	var b strings.Builder
	e.root.write(&b, precMin)
	return b.String()
}

// UnknownIDs returns the sorted, de-duplicated set of leaf license and exception
// ids in the expression that are NOT in SPDXListVersion's list. LicenseRef- and
// DocumentRef- identifiers are never reported (they are custom by definition).
// Empty means every id is recognized.
func (e *Expression) UnknownIDs() []string {
	seen := map[string]struct{}{}
	e.root.collectUnknown(seen)
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// --- AST ---------------------------------------------------------------------

// Operator precedence, loosest to tightest. write() parenthesizes a child whose
// precedence is looser than the context it is rendered in.
const (
	precMin  = iota // top-level / inside parens
	precOr          // OR
	precAnd         // AND
	precWith        // WITH
	precAtom        // id, id+, ref, ( ... )
)

type node interface {
	write(b *strings.Builder, ctx int)
	collectUnknown(seen map[string]struct{})
}

// licenseNode is a license id with optional "+" (or-later). canonical holds the
// official SPDX casing when known; when unknown it equals the as-written id.
type licenseNode struct {
	canonical string
	plus      bool
	known     bool
}

func (n *licenseNode) write(b *strings.Builder, _ int) {
	b.WriteString(n.canonical)
	if n.plus {
		b.WriteByte('+')
	}
}

func (n *licenseNode) collectUnknown(seen map[string]struct{}) {
	if !n.known {
		seen[n.canonical] = struct{}{}
	}
}

// refNode is a LicenseRef-/DocumentRef- custom identifier. It is always
// grammatically valid and never id-list-checked; its casing is preserved.
type refNode struct{ raw string }

func (n *refNode) write(b *strings.Builder, _ int)    { b.WriteString(n.raw) }
func (n *refNode) collectUnknown(map[string]struct{}) {}

// withNode is "<license> WITH <exception>".
type withNode struct {
	lic          node
	exception    string
	excCanonical string
	excKnown     bool
}

func (n *withNode) write(b *strings.Builder, ctx int) {
	wrap := ctx > precWith
	if wrap {
		b.WriteByte('(')
	}
	n.lic.write(b, precWith)
	b.WriteString(" WITH ")
	b.WriteString(n.excCanonical)
	if wrap {
		b.WriteByte(')')
	}
}

func (n *withNode) collectUnknown(seen map[string]struct{}) {
	n.lic.collectUnknown(seen)
	if !n.excKnown {
		seen[n.excCanonical] = struct{}{}
	}
}

// binNode is an AND or OR.
type binNode struct {
	op          string // "AND" or "OR"
	prec        int    // precAnd or precOr
	left, right node
}

func (n *binNode) write(b *strings.Builder, ctx int) {
	wrap := ctx > n.prec
	if wrap {
		b.WriteByte('(')
	}
	n.left.write(b, n.prec)
	b.WriteByte(' ')
	b.WriteString(n.op)
	b.WriteByte(' ')
	n.right.write(b, n.prec)
	if wrap {
		b.WriteByte(')')
	}
}

func (n *binNode) collectUnknown(seen map[string]struct{}) {
	n.left.collectUnknown(seen)
	n.right.collectUnknown(seen)
}

// --- parser ------------------------------------------------------------------

// tokenize splits an expression into "(", ")" and word tokens. Whitespace
// separates words; parentheses are always their own token even when adjacent to
// a word. The word charset (alnum, "-", ".", ":", "+") is a superset of what the
// grammar allows; the parser rejects misuse.
func tokenize(s string) []string {
	var toks []string
	var word strings.Builder
	flush := func() {
		if word.Len() > 0 {
			toks = append(toks, word.String())
			word.Reset()
		}
	}
	for _, r := range s {
		switch r {
		case '(', ')':
			flush()
			toks = append(toks, string(r))
		case ' ', '\t', '\n', '\r':
			flush()
		default:
			word.WriteRune(r)
		}
	}
	flush()
	return toks
}

type parser struct {
	toks []string
	pos  int
}

func (p *parser) atEnd() bool { return p.pos >= len(p.toks) }

func (p *parser) peek() string {
	if p.atEnd() {
		return ""
	}
	return p.toks[p.pos]
}

func (p *parser) next() string { t := p.peek(); p.pos++; return t }

func (p *parser) parseOr() (node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek() == "OR" {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &binNode{op: "OR", prec: precOr, left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (node, error) {
	left, err := p.parseWith()
	if err != nil {
		return nil, err
	}
	for p.peek() == "AND" {
		p.next()
		right, err := p.parseWith()
		if err != nil {
			return nil, err
		}
		left = &binNode{op: "AND", prec: precAnd, left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseWith() (node, error) {
	lic, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	if p.peek() != "WITH" {
		return lic, nil
	}
	// SPDX: "simple-expression WITH license-exception-id". The left of WITH must
	// be a bare license (id, id+, or LicenseRef), never a parenthesized group.
	if !isSimpleLicense(lic) {
		return nil, fmt.Errorf("%w: WITH must follow a simple license, not a group", ErrSyntax)
	}
	p.next() // consume WITH
	if p.atEnd() || isOperator(p.peek()) || p.peek() == "(" || p.peek() == ")" {
		return nil, fmt.Errorf("%w: WITH must be followed by an exception id", ErrSyntax)
	}
	excTok := p.next()
	if strings.HasSuffix(excTok, "+") || isRefToken(excTok) || !validIDChars(excTok) {
		return nil, fmt.Errorf("%w: invalid exception id %q", ErrSyntax, excTok)
	}
	canon, known := lookupException(excTok)
	return &withNode{lic: lic, exception: excTok, excCanonical: canon, excKnown: known}, nil
}

func (p *parser) parsePrimary() (node, error) {
	if p.atEnd() {
		return nil, fmt.Errorf("%w: expected a license, found end of input", ErrSyntax)
	}
	tok := p.peek()
	if tok == "(" {
		p.next()
		inner, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.peek() != ")" {
			return nil, fmt.Errorf("%w: missing closing parenthesis", ErrSyntax)
		}
		p.next()
		return inner, nil
	}
	if tok == ")" || isOperator(tok) {
		return nil, fmt.Errorf("%w: expected a license, found %q", ErrSyntax, tok)
	}
	p.next()
	return makeLeaf(tok)
}

func isOperator(tok string) bool { return tok == "AND" || tok == "OR" || tok == "WITH" }

func isRefToken(tok string) bool {
	return strings.HasPrefix(tok, "LicenseRef-") || strings.HasPrefix(tok, "DocumentRef-")
}

func isSimpleLicense(n node) bool {
	switch n.(type) {
	case *licenseNode, *refNode:
		return true
	}
	return false
}

// makeLeaf turns a word token into a licenseNode or refNode, validating shape.
func makeLeaf(tok string) (node, error) {
	if isRefToken(tok) {
		if !validRef(tok) {
			return nil, fmt.Errorf("%w: invalid license reference %q", ErrSyntax, tok)
		}
		return &refNode{raw: tok}, nil
	}
	plus := strings.HasSuffix(tok, "+")
	id := tok
	if plus {
		id = id[:len(id)-1]
	}
	if id == "" || !validIDChars(id) {
		return nil, fmt.Errorf("%w: invalid license id %q", ErrSyntax, tok)
	}
	canon, known := lookupLicense(id)
	return &licenseNode{canonical: canon, plus: plus, known: known}, nil
}

// validRef checks the LicenseRef-/DocumentRef- grammar:
//
//	["DocumentRef-"idstring":"] "LicenseRef-"idstring
//
// idstring = 1*(ALPHA / DIGIT / "-" / "."). Casing is preserved (these are
// case-sensitive, unlike standard SPDX ids).
func validRef(tok string) bool {
	rest := tok
	if after, ok := strings.CutPrefix(rest, "DocumentRef-"); ok {
		doc, lic, found := strings.Cut(after, ":")
		if !found || !validIDChars(doc) {
			return false
		}
		rest = lic
	}
	after, ok := strings.CutPrefix(rest, "LicenseRef-")
	return ok && validIDChars(after)
}

func validIDChars(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '.':
		default:
			return false
		}
	}
	return true
}

// lookupLicense / lookupException map a case-insensitive id to its canonical
// SPDX casing. The second return is false when the id is not in the bundled
// list (SPDXListVersion). Data lives in the generated spdx_data.go.
func lookupLicense(id string) (canonical string, known bool) {
	if c, ok := spdxLicenseIDs[strings.ToLower(id)]; ok {
		return c, true
	}
	return id, false
}

func lookupException(id string) (canonical string, known bool) {
	if c, ok := spdxExceptionIDs[strings.ToLower(id)]; ok {
		return c, true
	}
	return id, false
}
