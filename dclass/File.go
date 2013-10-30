package dclass

type File struct {
	Classes map[string]*Class  // a map of class-name to distributed classes declared in the file
	Structs map[string]*Struct // a map of class-name to distributed structs declared in the file

	keywords // implements KeywordList
}

// Hash returns a hash of the file's structure. Hash implements the Hashable interface.
// TODO: Implement
func (f *File) Hash() uint64 {
	return 0
}

// NewStruct returns a new struct initialized with a name and unique index within the dclass file
func (f *File) NewStruct(name string) *Struct {
	s := new(Struct)
	s.dcf = f
	s.name = name
	//c.index = f.nextIndex()
	return s
}

// NewStruct returns a new class initialized with a name and unique index within the dclass file
func (f *File) NewClass(name string) *Class {
	c := new(Class)
	c.dcf = f
	c.name = name
	//c.index = f.nextIndex()
	return c
}

// addField is called by classes and structs to add a new field to the file
// returns the unique index of the field
// TODO: Implement
func (f *File) addField(field *Field) int {
	return 0
}
