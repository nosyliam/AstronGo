package main

import (
	"github.com/alecthomas/participle"
	"github.com/davecgh/go-spew/spew"
)

const dcf = ``

type KeywordType struct {
	Name string `@Ident`
}

type KeywordList struct {
	Keywords []string `{ @Ident }`
}

type Range struct {
	Lo *int64 `'(' @Int '-' `
	Hi *int64 `@Int ')'`
}

type ArrayRange struct {
	Lo *int64 `@Int`
	Hi *int64 `[ '-' @Int ]`
}

type ArrayValue struct {
	String     *string `@String`
	Negative   bool    `| [ @"-" ]`
	Int        *int64  `( '-' @Int | @Int )`
	Multiplier *int64  `[ "*" @Int ]`
}

type ArrayBounds struct {
	Array           bool        `@'['`
	ArrayConstraint *ArrayRange `[ @@ ] ']'`
}

type DefaultValue struct {
	Array   []*ArrayValue `( '[' @@ { "," @@ } ']' ) |`
	Integer *int          `( @Int ) |`
	Float   *float64      `( @Float ) |`
	String  *string       `( @String )`
}

type CharParameter struct {
	Type        string         `@"char"`
	ArrayPrefix []*ArrayBounds `@@ @@`
	Identifier  *string        `[ @Ident ]`
	ArraySuffix []*ArrayBounds `@@ @@`
	Default     *DefaultValue  `[ '=' @@ ]`
}

type IntTransform struct {
	Operator string `@( "%" | "*" | "+" | "-" | "/" )`
	Value    int    `@Int`
}

type IntParameter struct {
	Type        string          `@( "int8" | "int16" | "int32" | "int64" | "uint8" | "uint16" | "uint32" | "uint64" )`
	Transforms  []*IntTransform `[ { @@ } ]`
	Constraint  *Range          `[ @@ ]`
	ArrayPrefix []*ArrayBounds  `[ { @@ } ]`
	Identifier  *string         `[ @Ident ]`
	ArraySuffix []*ArrayBounds  `[ { @@ } ]`
	Default     *DefaultValue   `[ '=' @@ ]`
}

type FloatParameter struct {
	Type       string          `@"float64"`
	Transforms []*IntTransform `[ @@ { @@ } ]`
	Constraint *Range          `[ @@ ]`
	Identifier *string         `[ @Ident ]`
	Default    *DefaultValue   `[ '=' @@ ]`
}

type SizedParameter struct {
	Type        string         `@( "string" | "blob" )`
	Constraint  *Range         `[ @@ ]`
	ArrayPrefix []*ArrayBounds `[ { @@ } ]`
	Identifier  *string        `[ @Ident ]`
	ArraySuffix []*ArrayBounds `[ { @@ } ]`
	Default     *DefaultValue  `[ '=' @@ ]`
}

type AmbiguousParameter struct {
	Type            string         `@Ident`
	Identifier      string         `[ @Ident ]`
	ArrayConstraint []*ArrayBounds `[ { @@ } ]`
	Default         *DefaultValue  `[ '=' @@ ]`
}

type Parameter struct {
	Char  *CharParameter      `@@`
	Int   *IntParameter       `| @@`
	Float *FloatParameter     `| @@`
	Sized *SizedParameter     `| @@`
	Typed *AmbiguousParameter `| @@`
}

type AtomicField struct {
	Name       string       `@Ident`
	Parameters []*Parameter `'(' [ @@ ] { "," @@ } ')'`
	Keywords   *KeywordList `[ @@ ]`
}

type MolecularField struct {
	Name   string   `@Ident`
	Fields []string `':' @Ident { ',' @Ident }`
}

type ParameterField struct {
	Parameter *Parameter   `@@`
	Keywords  *KeywordList `[ @@ ]`
}

type FieldDecl struct {
	Atomic    *AtomicField    `@@`
	Molecular *MolecularField `| @@`
	Parameter *ParameterField `| @@`
}

type ClassType struct {
	Name         string       `"dclass" @Ident`
	Parents      []string     `[ ':' @Ident { ',' @Ident } ]`
	Declarations []*FieldDecl `'{' { @@ ';' } '}' ';'`
}

type StructType struct {
	Name       string       `"struct" @Ident`
	Parameters []*Parameter `'{' { @@ ';' } '}' ';'`
}

type Typedef struct {
	Name string // `"typedef "`
}

type TypeDecl struct {
	Keyword *KeywordType `"keyword" @@`
	Struct  *StructType  `| @@`
	Class   *ClassType   `| @@`
}

type DCFile struct {
	Declarations []*TypeDecl `{ @@ }`
}

func main() {
	parser, err := participle.Build(&DCFile{}, participle.UseLookahead(16))
	if err != nil {
		panic(err)
	}

	dc := &DCFile{}
	err = parser.ParseString(dcf, dc)
	spew.Dump(dc)
}
