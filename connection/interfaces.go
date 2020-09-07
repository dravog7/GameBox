package connection

import "time"

/*

Connection and Connection Factory basic interface

TODO - Add support for JSON and binary messages

*/

//Connection - Defines client Connections
type Connection interface {
	Listen(func(Connection, string, string)) string    //add listeners for messages
	Reconnect(Connection) Connection                   //reconnect a connection
	setUID(string)                                     //set uid on reconnection
	listenTo(func(Connection, string, string), string) //listenTo for setting listener with uid provided
	Remove(string) error                               //remove a listener
	Send(string) error                                 //send string
	SendJSON(msg interface{}) error                    //send JSON
	GetParams() map[string]string                      //get Params generated during Setup
	Close() error                                      //Close connection
	IsClosed() bool                                    //Check if Closed
	String() string                                    //uid String of connection
	GetPing() time.Duration                            //Get the ping of connection
}

//Factory - Defines Factory which generates Connection objects
type Factory interface {
	New(func(Connection, map[string]string))
}
