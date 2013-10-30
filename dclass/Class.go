package dclass

var definedKeywords = keywords{
	"required",
	"ram",
	"broadcast",
	"clrecv",
	"clsend",
	"ownrecv",
	"ownsend",
	"airecv",
	"db",
}

// a typeBase is a parent type that the Class and Struct types should extend
type typeBase struct {
	dcf   *File  // file this type is associated with
	name  string // name of the type
	index int    // the unique index of the type within the dclass file
}

type Class struct {
	typeBase // inherits from typeBase
}

// Hash returns a hash of the class's structure. Hash implements the Hashable interface.
// TODO: Implement
func (c *Class) Hash() uint64 {
	return 0
}

type Struct struct {
	typeBase // inherits from typeBase
}

// Hash returns a hash of the struct's structure. Hash implements the Hashable interface.
// TODO: Implement
func (s *Struct) Hash() uint64 {
	return 0
}