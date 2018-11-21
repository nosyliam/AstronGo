package parse

type KeywordType struct {
	Name string `@Ident`
}

type KeywordList struct {
	Keywords []string `{ @Ident }`
}

type Number struct {
	Negative bool `[ @"-" ]`
	Value *int64 `@Int`
}

type Range struct {
	Lo *Number `'(' @@ [ '-' `
	Hi *Number `@@ ] ')'`
}

type ArrayRange struct {
	Lo *int64 `@Int`
	Hi *int64 `[ '-' @Int ]`
}

type ArrayValue struct {
	String     *string `@String`
	Number     *Number `| @@`
	Multiplier *int64  `[ "*" @Int ]`
}

type ArrayBounds struct {
	Array           bool        `@'['`
	ArrayConstraint *ArrayRange `[ @@ ] ']'`
}

type DefaultValue struct {
	Array        bool          `( @'['`
	ArrayDefault []*ArrayValue `[ @@ ] { "," @@ } ']' ) |`
	Negative     bool          `( [ @"-" ]`
	Integer      *int          `@Int`
	Float        *float64      `| @Float )`
	String       *string       `| @String`
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
	Declarations []*FieldDecl `'{' { @@ ';' } '}' [ ';' ]`
}

type StructType struct {
	Name       string       `"struct" @Ident`
	Parameters []*Parameter `'{' { @@ ';' } '}' [ ';' ]`
}

type TypeCapture struct {
	Name       string          `@Ident`
	Transforms []*IntTransform `[ { @@ } ]`
	Constraint *Range          `[ @@ ]`
	Bounds     []*ArrayBounds  `[ { @@ } ]`
}

type Typedef struct {
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
	Keyword *KeywordType `"keyword" @@`
	Import  *Import      `| @@`
	Typedef *Typedef     `| @@`
	Struct  *StructType  `| @@`
	Class   *ClassType   `| @@`
}

type DCFile struct {
	Declarations []*TypeDecl `{ @@ }`
}
