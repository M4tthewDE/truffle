package twitch

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type reader struct {
	eventChan chan Event
	partChan  chan bool
}

func newReader(eventChan chan Event) reader {
	return reader{
		eventChan: eventChan,
		partChan:  make(chan bool),
	}
}

type websocketMessage struct {
	messageType int
	data        []byte
	err         error
}

func (r *reader) read(auth Authentication, condition Condition) {
	u := url.URL{Scheme: "wss", Host: "eventsub.wss.twitch.tv", Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer c.Close()

	msgChan := make(chan websocketMessage)

	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Printf("Parted %s\n", condition.BroadcasterUserId)
				close(msgChan)
				return
			default:
				messageType, message, err := c.ReadMessage()
				msgChan <- websocketMessage{
					messageType: messageType,
					data:        message,
					err:         err,
				}
			}

		}
	}(ctx)

	for {
		select {
		case wsMsg, ok := <-msgChan:
			if !ok {
				cancel()
				return
			}
			if wsMsg.err != nil {
				log.Println(err)
				cancel()
			}

			var msg Message
			err = json.Unmarshal(wsMsg.data, &msg)
			if err != nil {
				log.Println(err)
				cancel()
			}

			if msg.Metadata.MessageType == "session_welcome" {
				_, err := createMessageSub(auth, msg.Payload.Session.Id, condition)
				if err != nil {
					log.Println(err)
					cancel()
				}
			}

			if msg.Metadata.MessageType == "session_reconnect" {
				// TODO: implement reconnect logic
				log.Println("session_reconnect")
			}

			if msg.Metadata.MessageType == "revocation" {
				// TODO: what do we do in this case?
				log.Println("revocation")
			}

			r.eventChan <- msg.Payload.Event
		case <-r.partChan:
			log.Printf("Parting %s", condition.BroadcasterUserId)
			cancel()
		}
	}

}

type joinRequest struct {
	id     uuid.UUID
	wsChan chan Event
	auth   Authentication
	cond   Condition
}

type partRequest struct {
	id        uuid.UUID
	channelId string
}

type readerManager struct {
	readers   map[string]reader
	joinChan  chan joinRequest
	partChan  chan partRequest
	eventChan chan Event
	wsChans   map[string]map[uuid.UUID]chan Event
}

func newReaderManager() readerManager {
	return readerManager{
		readers:   make(map[string]reader),
		joinChan:  make(chan joinRequest),
		partChan:  make(chan partRequest),
		eventChan: make(chan Event),
		wsChans:   make(map[string]map[uuid.UUID]chan Event),
	}
}

func (r *readerManager) run() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case join := <-r.joinChan:
			channelId := join.cond.BroadcasterUserId
			_, connected := r.readers[channelId]
			if connected {
				log.Printf("Already connected to %s\n", channelId)
				r.wsChans[channelId][join.id] = join.wsChan
			} else {
				log.Printf("Connecting to %s\n", channelId)
				r.wsChans[channelId] = map[uuid.UUID]chan Event{join.id: join.wsChan}
				reader := newReader(r.eventChan)
				go reader.read(join.auth, join.cond)
				r.readers[channelId] = reader
			}
		case part := <-r.partChan:
			log.Printf("Removing connection to %s for %s\n", part.channelId, part.id)
			delete(r.wsChans[part.channelId], part.id)
		case event := <-r.eventChan:
			for _, ws := range r.wsChans[event.BroadcasterUserId] {
				ws <- event
			}
		case <-ticker.C:
			for channelId, wsChans := range r.wsChans {
				log.Printf("%d frontend(s) listening to %s\n", len(wsChans), channelId)
				if len(wsChans) == 0 {
					log.Printf("Cleaning up %s\n", channelId)
					r.readers[channelId].partChan <- true
					delete(r.wsChans, channelId)
					delete(r.readers, channelId)
				}
			}
		default:
		}
	}
}

var (
	rm readerManager
)

func Init() {
	rm = newReaderManager()
	go rm.run()
}

func Join(auth Authentication, cond Condition, wsChan chan Event) uuid.UUID {
	id := uuid.New()
	fConn := joinRequest{
		id:     id,
		wsChan: wsChan,
		auth:   auth,
		cond:   cond,
	}
	rm.joinChan <- fConn

	return id
}

func Part(channelId string, id uuid.UUID) {
	partRequest := partRequest{
		id:        id,
		channelId: channelId,
	}

	rm.partChan <- partRequest
}
