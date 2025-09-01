package types

type ParseState string

const (
	StateRequestLine ParseState = "RequestLine"
	StateHeader      ParseState = "Header"
	StateBody        ParseState = "Body"
	StateDone        ParseState = "Done"
)
