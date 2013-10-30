package dclass

import "bytes"

// A Field is a member (Parameter), function (AtomicField), or composite function (MolecularField)
// of a dclass Class object. Field inherits from Hashable, requiring concrete fields to implement Hash.
type Field interface {
	Hashable

	// Name returns the name of this field parsed from a file
	Name() string

	// Number returns the index of this field which is unqiue within its dclass File
	Number() int

	// NestedFields returns the nested fields of this field. Nested fields typically represent a
	// function's arguments (the Parameters of an AtomicField) or the components of a composite
	// function (any number of components of a MolecularField).
	NestedFields() []Field

	// File returns the dclass File this field is associated with
	File() *File

	// DefaultValue returns the default value specified in the dclass File,
	// or the null value if no default was specified (typically 0).
	DefaultValue() bytes.Buffer

	// HasDefaultValue returns whether a default value was specified in the dclass File.
	HasDefaultValue() bool

	// The IsFoo methods return whether the field has the keyword "foo". While some fields may imply
	// other fields in terms of behavior, these methods only return true if that keyword was explicitly
	// set within the dclass File.
	IsRequired() bool
	IsRam() bool
	IsBroadcast() bool
	IsClrecv() bool
	IsClsend() bool
	IsOwnrecv() bool
	IsOwnsend() bool
	IsAirecv() bool
	IsDb() bool

	// FormatData accepts a blob that represents the packed data for this field and returns a string
	// formatting it for human consumption.
	FormatData(data bytes.Buffer, showFieldNames bool) string

	// ParseString accepts a human readable string (the output of FormatData) and returns a buffer
	// that has the packed data for this field.
	ParseString(s string) (data bytes.Buffer, err error)
}

// a fieldBase is a parent type which other implementations of the Field interface can extend
type fieldBase struct {
	dcf      *File  // file this type is associated with
	name     string // name of the field
	index    int    // the unique index of the type within the dclass file
	keywords        // implements KeywordList
}

type Parameter struct {
	fieldBase // inherits from fieldBase

	dataTyp DataType
	defVal  bytes.Buffer
}
type AtomicField struct {
	fieldBase // inherits from fieldBase
}
type MolecularField struct {
	fieldBase // inherits from fieldBase
}

// AddField creates a new field and adds it to the object.
// Atomic fields can only accept a "Parameter" field type.
// TODO: Implement
func (f *AtomicField) AddField(name, typ string) Field {
	return nil
}

// AddField creates a new field and adds it to the object.
// Molecular fields can accept "Parameter" and "AtomicField" field types.
// TODO: Implement
func (f *MolecularField) AddField(name, typ string) Field {
	return nil
}
