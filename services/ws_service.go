package services

import (
	"encoding/json"
	"errors"
	"express-wallet-trending/models"
	url2 "net/url"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type PayloadHandler interface {
	Handle() error
}

type WsService struct {
	url  url2.URL
	conn *websocket.Conn
	PayloadHandler
}

func NewWsConn(scheme string, host string, path string) (*WsService, error) {
	h := &WsService{}
	url := url2.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
	h.url = url
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return nil, err
	}
	h.conn = conn
	h.conn.SetCloseHandler(h.CloseHandler)
	return h, nil
}

func (h *WsService) ReConn() error {
	conn, _, err := websocket.DefaultDialer.Dial(h.url.String(), nil)
	if err != nil {
		return err
	}
	h.conn = conn
	h.conn.SetCloseHandler(h.CloseHandler)
	return nil
}

func (h *WsService) Close() error {
	return h.conn.Close()
}

func (h *WsService) CloseHandler(code int, text string) error {
	if err := h.ReConn(); err != nil {
		return err
	}
	return errors.New(text)
}

func (h *WsService) Handle(producer *KafkaProducer) error {
	for {
		_, msg, err := h.conn.ReadMessage()
		if err != nil {
			if connErr := h.ReConn(); connErr != nil {
				log.Errorln("websocket reconnect error: %v", connErr)
				return err
			}
		}
		bnPayload := &models.BnPayload{}
		err = json.Unmarshal(msg, bnPayload)
		if err != nil {
			log.Errorln("WsHandler unmarshal message error:", err)
			continue
		}
		key := bnPayload.Symbol + bnPayload.EventType
		producer.AsyncSendMessage(key, string(msg))
	}
}
