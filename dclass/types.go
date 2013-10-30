package dclass

// A DataType declares the type of data stored by a Parameter.
type DataType int

const (
	InvalidType DataType = iota
	Int8Type
	Int16Type
	Int32Type
	Int64Type
	Uint8Type
	Uint16Type
	Uint32Type
	Uint64Type
	FloatType
	StringType
	BlobType
	CharType
	StructType
	ArrayType
)

// An Error is a dclass package specific error
type Error string

// implements Error interface
func (err Error) Error() string {
	return string(err)
}

// implements Stringer interface
func (err Error) String() string {
	return string(err)
}

func runtimeError(msg string) Error {
	return Error("runtime error: " + msg)
}

type Hashable interface {
	Hash() uint64
}

// A KeywordList is any dctype that has an associated keyword list.  The most common KeywordLists
// are: a File object with its list of declared keywords and a Field with its list of enabled keywords.
type KeywordList interface {
	// AddKeyword adds the keyword argument to the set of keywords in the list.
	AddKeyword(keyword string)

	// AddKeywords performs a union of the KeywordList argument into this KeywordList.
	AddKeywords(list KeywordList)

	// CompareKeywords compares two KeywordLists and returns true if they contain the same set of
	// keywords. Order does not matter.
	CompareKeywords(list KeywordList) bool

	// HasKeyword returns whether the keyword argument exists in the list.
	HasKeyword(keyword string) bool

	// Keywords returns the list of keywords as a slice
	Keywords() []string

	// NumKeywords returns the length of the keyword list
	NumKeywords() int
}

// type keywords is a string slice satisfying the KeywordList interface.
type keywords []string

// implementing KeywordList
func (k keywords) AddKeyword(keyword string) {
	if !k.HasKeyword(keyword) {
		k = append(k, keyword)
	}
}

// implementing KeywordList
func (k keywords) AddKeywords(list KeywordList) {
	for _, keyword := range list.Keywords() {
		k.AddKeyword(keyword)
	}
}

// implementing KeywordList
func (k keywords) CompareKeywords(list KeywordList) bool {
	if len(k) != len(list.Keywords()) {
		return false
	}
	for _, keyword := range k {
		if !list.HasKeyword(keyword) {
			return false
		}
	}
	return true
}

// implementing KeywordList
func (k keywords) HasKeyword(keyword string) bool {
	for _, word := range k {
		if keyword == word {
			return true
		}
	}
	return false
}

// implementing KeywordList
func (k keywords) Keywords() []string {
	return []string(k)
}

// implementing KeywordList
func (k keywords) NumKeywords() int {
	return len(k)
}
