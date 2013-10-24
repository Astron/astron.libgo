package dclass

type File struct {
	Classes map[string]Class  // a map of class-name to distributed classes declared in the file
	Structs map[string]Struct // a map of class-name to distributed structs declared in the file

	keywords // implements KeywordList
}

// Hash returns a hash of the file's structure. Hash implements the Hashable interface.
// TODO: Implement
func (f *File) Hash() uint64 {
	return 0
}
