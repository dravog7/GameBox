package connection

import "github.com/google/uuid"

type InMessage struct {
	MessageType int
	Message     string
}

type Subscription struct {
	subscribers map[uuid.UUID]chan *InMessage
}

func (conn *Subscription) Subscribe(subscriber chan *InMessage) uuid.UUID {
	id := uuid.New()
	conn.subscribers[id] = subscriber
	return id
}

func (conn *Subscription) Unsubscribe(id uuid.UUID) error {
	delete(conn.subscribers, id)
	return nil
}

func (conn *Subscription) emit(msgType int, msg string) {
	message := &InMessage{msgType, msg}
	for _, subscriber := range conn.subscribers {
		subscriber <- message
	}
}

func (conn *Subscription) closeSubscription() {
	for _, subscriber := range conn.subscribers {
		close(subscriber)
	}
}
