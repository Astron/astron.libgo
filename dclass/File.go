package dclass

type File struct {
	Classes []Type  // a list of classes and structs associated with the file
	Fields  []Field // a list of fields associated with the file

	ClassByName map[string]Type // a map of class names to classes and structs

	keywords // implements KeywordList
}

// Hash returns a hash of the file's structure. Hash implements the Hashable interface.
// TODO: Implement
func (f *File) Hash() uint64 {
	return 0
}

// AddType returns a new Type initialized with a name and unique index within the dclass file.
// The typ argument can be either "class" or "struct".
func (f *File) AddType(name, typ string) Type {
	switch typ {
	case "class":
		c := new(Class)
		c.dcf = f
		c.name = name
		c.index = len(f.Classes)
		f.Classes = append(f.Classes, c)
		f.ClassByName[name] = c
		return c
	case "struct":
		s := new(Struct)
		s.dcf = f
		s.name = name
		s.index = len(f.Classes)
		f.Classes = append(f.Classes, s)
		f.ClassByName[name] = s
		return s
	default:
		return nil
	}
}

// addField is called by classes and structs to add a new field to the file
// returns the unique index of the field
// TODO: Implement
func (f *File) addField(field *Field) int {
	return 0
}
