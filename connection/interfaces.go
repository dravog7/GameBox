package connection

/*

Connection and Connection Factory basic interface

TODO - Add support for JSON and binary messages

*/

//Connection - Defines client Connections
type Connection interface {
	Listen(func(Connection, string, string)) string
	Remove(string) error
	Send(string) error
	Close() error
	String() string
}

//Factory - Defines Factory which generates Connection objects
type Factory interface {
	New(func(Connection, map[string]string))
}
