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

type Class struct {
}

// Hash returns a hash of the class's structure. Hash implements the Hashable interface.
// TODO: Implement
func (c *Class) Hash() uint64 {
	return 0
}

type Struct struct {
}

// Hash returns a hash of the struct's structure. Hash implements the Hashable interface.
// TODO: Implement
func (s *Struct) Hash() uint64 {
	return 0
}
