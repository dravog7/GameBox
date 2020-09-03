package connection

//Connection - Defines client Connections
type Connection interface {
	Listen(func(Connection, string, string))
	Send(string) error
	Close() error
	String() string
}

//Factory - Defines Factory which generates Connection objects
type Factory interface {
	New(func(Connection, map[string]string))
}
