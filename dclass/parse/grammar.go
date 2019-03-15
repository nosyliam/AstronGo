package parse

import "github.com/alecthomas/participle/lexer"

type KeywordType struct {
	Pos  lexer.Position
	Name string `@Ident`
}

type KeywordList struct {
	Pos      lexer.Position
	Keywords []string `{ @Ident }`
}

type Number struct {
	Pos      lexer.Position
	Negative bool   `[ @"-" ]`
	Value    *int64 `@Int`
}

type Range struct {
	Pos lexer.Position
	Lo  *Number `'(' @@ [ '-' `
	Hi  *Number `@@ ] ')'`
}

type ArrayRange struct {
	Pos lexer.Position
	Lo  *int64 `@Int`
	Hi  *int64 `[ '-' @Int ]`
}

type ArrayValue struct {
	Pos        lexer.Position
	String     *string `@String`
	Number     *Number `| @@`
	Multiplier *int64  `[ "*" @Int ]`
}

type ArrayBounds struct {
	Pos             lexer.Position
	Array           bool        `@'['`
	ArrayConstraint *ArrayRange `[ @@ ] ']'`
}

type DefaultValue struct {
	Pos          lexer.Position
	Array        bool          `( @'['`
	ArrayDefault []*ArrayValue `[ @@ ] { "," @@ } ']' ) |`
	Negative     bool          `( [ @"-" ]`
	Integer      *int          `@Int`
	Float        *float64      `| @Float )`
	String       *string       `| @String`
}

type CharParameter struct {
	Pos         lexer.Position
	Type        string         `@"char"`
	ArrayPrefix []*ArrayBounds `@@ @@`
	Identifier  *string        `[ @Ident ]`
	ArraySuffix []*ArrayBounds `@@ @@`
	Default     *DefaultValue  `[ '=' @@ ]`
}

type IntTransform struct {
	Pos      lexer.Position
	Operator string `@( "%" | "*" | "+" | "-" | "/" )`
	Value    int    `@Int`
}

type IntParameter struct {
	Pos         lexer.Position
	Type        string          `@( "int8" | "int16" | "int32" | "int64" | "uint8" | "uint16" | "uint32" | "uint64" )`
	Transforms  []*IntTransform `[ { @@ } ]`
	Constraint  *Range          `[ @@ ]`
	ArrayPrefix []*ArrayBounds  `[ { @@ } ]`
	Identifier  *string         `[ @Ident ]`
	ArraySuffix []*ArrayBounds  `[ { @@ } ]`
	Default     *DefaultValue   `[ '=' @@ ]`
}

type FloatParameter struct {
	Pos        lexer.Position
	Type       string          `@"float64"`
	Transforms []*IntTransform `[ @@ { @@ } ]`
	Constraint *Range          `[ @@ ]`
	Identifier *string         `[ @Ident ]`
	Default    *DefaultValue   `[ '=' @@ ]`
}

type SizedParameter struct {
	Pos         lexer.Position
	Type        string         `@( "string" | "blob" )`
	Constraint  *Range         `[ @@ ]`
	ArrayPrefix []*ArrayBounds `[ { @@ } ]`
	Identifier  *string        `[ @Ident ]`
	ArraySuffix []*ArrayBounds `[ { @@ } ]`
	Default     *DefaultValue  `[ '=' @@ ]`
}

type AmbiguousParameter struct {
	Pos             lexer.Position
	Type            string         `@Ident`
	Identifier      string         `[ @Ident ]`
	ArrayConstraint []*ArrayBounds `[ { @@ } ]`
	Default         *DefaultValue  `[ '=' @@ ]`
}

type Parameter struct {
	Pos   lexer.Position
	Char  *CharParameter      `@@`
	Int   *IntParameter       `| @@`
	Float *FloatParameter     `| @@`
	Sized *SizedParameter     `| @@`
	Typed *AmbiguousParameter `| @@`
}

type AtomicField struct {
	Pos        lexer.Position
	Name       string       `@Ident`
	Parameters []*Parameter `'(' [ @@ ] { "," @@ } ')'`
	Keywords   *KeywordList `[ @@ ]`
}

type MolecularField struct {
	Pos    lexer.Position
	Name   string   `@Ident`
	Fields []string `':' @Ident { ',' @Ident }`
}

type ParameterField struct {
	Pos       lexer.Position
	Parameter *Parameter   `@@`
	Keywords  *KeywordList `[ @@ ]`
}

type FieldDecl struct {
	Pos       lexer.Position
	Atomic    *AtomicField    `@@`
	Molecular *MolecularField `| @@`
	Parameter *ParameterField `| @@`
}

type ClassType struct {
	Pos          lexer.Position
	Name         string       `"dclass" @Ident`
	Parents      []string     `[ ':' @Ident { ',' @Ident } ]`
	Declarations []*FieldDecl `'{' { @@ ';' } '}' [ ';' ]`
}

type StructType struct {
	Pos        lexer.Position
	Name       string       `"struct" @Ident`
	Parameters []*Parameter `'{' { @@ ';' } '}' [ ';' ]`
}

type TypeCapture struct {
	Pos        lexer.Position
	Name       string          `@Ident`
	Transforms []*IntTransform `[ { @@ } ]`
	Constraint *Range          `[ @@ ]`
	Bounds     []*ArrayBounds  `[ { @@ } ]`
}

type Typedef struct {
	Pos  lexer.Position
	Base *TypeCapture `"typedef" @@`
	Type *TypeCapture `@@ ';'`
}

type Import struct {
	Path              []string `"from" @Ident { '.' @Ident }`
	PathDenominators  []string `[ { '/' @Ident } ]`
	Class             string   `"import" ( @Ident | '*' )`
	ClassDenominators []string `[ { '/' @Ident } ]`
}

type TypeDecl struct {
	Pos     lexer.Position
	Keyword *KeywordType `"keyword" @@`
	Import  *Import      `| @@`
	Typedef *Typedef     `| @@`
	Struct  *StructType  `| @@`
	Class   *ClassType   `| @@`
}

type DCFile struct {
	Declarations []*TypeDecl `{ @@ }`
}
