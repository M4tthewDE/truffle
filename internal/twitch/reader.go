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

type Reader struct {
	eventChan chan Event
	partChan  chan bool
}

func NewReader(eventChan chan Event) Reader {
	return Reader{
		eventChan: eventChan,
		partChan:  make(chan bool),
	}
}

type websocketMessage struct {
	messageType int
	data        []byte
	err         error
}

func (r *Reader) read(auth Authentication, condition Condition) {
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
	Id     uuid.UUID
	WsChan chan Event
	Auth   Authentication
	Cond   Condition
}

type partRequest struct {
	Id        uuid.UUID
	ChannelId string
}

type ReaderManager struct {
	readers   map[string]Reader
	joinChan  chan joinRequest
	partChan  chan partRequest
	eventChan chan Event
	wsChans   map[string]map[uuid.UUID]chan Event
}

func newReaderManager() ReaderManager {
	return ReaderManager{
		readers:   make(map[string]Reader),
		joinChan:  make(chan joinRequest),
		partChan:  make(chan partRequest),
		eventChan: make(chan Event),
		wsChans:   make(map[string]map[uuid.UUID]chan Event),
	}
}

func (r *ReaderManager) run() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case join := <-r.joinChan:
			channelId := join.Cond.BroadcasterUserId
			_, connected := r.readers[channelId]
			if connected {
				log.Printf("Already connected to %s\n", channelId)
				r.wsChans[channelId][join.Id] = join.WsChan
			} else {
				log.Printf("Connecting to %s\n", channelId)
				r.wsChans[channelId] = map[uuid.UUID]chan Event{join.Id: join.WsChan}
				reader := NewReader(r.eventChan)
				go reader.read(join.Auth, join.Cond)
				r.readers[channelId] = reader
			}
		case part := <-r.partChan:
			log.Printf("Removing connection to %s for %s\n", part.ChannelId, part.Id)
			delete(r.wsChans[part.ChannelId], part.Id)
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
	readerManager ReaderManager
)

func Init() {
	readerManager = newReaderManager()
	go readerManager.run()
}

func Join(auth Authentication, cond Condition, wsChan chan Event) uuid.UUID {
	id := uuid.New()
	fConn := joinRequest{
		Id:     id,
		WsChan: wsChan,
		Auth:   auth,
		Cond:   cond,
	}
	readerManager.joinChan <- fConn

	return id
}

func Part(channelId string, id uuid.UUID) {
	partRequest := partRequest{
		Id:        id,
		ChannelId: channelId,
	}

	readerManager.partChan <- partRequest
}
