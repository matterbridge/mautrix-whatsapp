// mautrix-whatsapp - A Matrix-WhatsApp puppeting bridge.
// Copyright (C) 2019 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package database

import (
	"bytes"
	"database/sql"
	"encoding/json"

	waProto "github.com/matterbridge/go-whatsapp/binary/proto"

	log "maunium.net/go/maulogger/v2"

	"github.com/matterbridge/mautrix-whatsapp/types"
)

type MessageQuery struct {
	db  *Database
	log log.Logger
}

func (mq *MessageQuery) CreateTable() error {
	_, err := mq.db.Exec(`CREATE TABLE IF NOT EXISTS message (
		chat_jid      VARCHAR(25),
		chat_receiver VARCHAR(25),
		jid           VARCHAR(255),
		mxid          VARCHAR(255) NOT NULL UNIQUE,
		sender        VARCHAR(25)  NOT NULL,
		content       BLOB         NOT NULL,

		PRIMARY KEY (chat_jid, chat_receiver, jid),
		FOREIGN KEY (chat_jid, chat_receiver) REFERENCES portal(jid, receiver)
	)`)
	return err
}

func (mq *MessageQuery) New() *Message {
	return &Message{
		db:  mq.db,
		log: mq.log,
	}
}

func (mq *MessageQuery) GetAll(chat PortalKey) (messages []*Message) {
	rows, err := mq.db.Query("SELECT * FROM message WHERE chat_jid=? AND chat_receiver=?", chat.JID, chat.Receiver)
	if err != nil || rows == nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		messages = append(messages, mq.New().Scan(rows))
	}
	return
}

func (mq *MessageQuery) GetByJID(chat PortalKey, jid types.WhatsAppMessageID) *Message {
	return mq.get("SELECT * FROM message WHERE chat_jid=? AND chat_receiver=? AND jid=?", chat.JID, chat.Receiver, jid)
}

func (mq *MessageQuery) GetByMXID(mxid types.MatrixEventID) *Message {
	return mq.get("SELECT * FROM message WHERE mxid=?", mxid)
}

func (mq *MessageQuery) get(query string, args ...interface{}) *Message {
	row := mq.db.QueryRow(query, args...)
	if row == nil {
		return nil
	}
	return mq.New().Scan(row)
}

type Message struct {
	db  *Database
	log log.Logger

	Chat    PortalKey
	JID     types.WhatsAppMessageID
	MXID    types.MatrixEventID
	Sender  types.WhatsAppID
	Content *waProto.Message
}

func (msg *Message) Scan(row Scannable) *Message {
	var content []byte
	err := row.Scan(&msg.Chat.JID, &msg.Chat.Receiver, &msg.JID, &msg.MXID, &msg.Sender, &content)
	if err != nil {
		if err != sql.ErrNoRows {
			msg.log.Errorln("Database scan failed:", err)
		}
		return nil
	}

	msg.decodeBinaryContent(content)

	return msg
}

func (msg *Message) decodeBinaryContent(content []byte) {
	msg.Content = &waProto.Message{}
	reader := bytes.NewReader(content)
	dec := json.NewDecoder(reader)
	err := dec.Decode(msg.Content)
	if err != nil {
		msg.log.Warnln("Failed to decode message content:", err)
	}
}

func (msg *Message) encodeBinaryContent() []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(msg.Content)
	if err != nil {
		msg.log.Warnln("Failed to encode message content:", err)
	}
	return buf.Bytes()
}

func (msg *Message) Insert() {
	_, err := msg.db.Exec("INSERT INTO message VALUES (?, ?, ?, ?, ?, ?)", msg.Chat.JID, msg.Chat.Receiver, msg.JID, msg.MXID, msg.Sender, msg.encodeBinaryContent())
	if err != nil {
		msg.log.Warnfln("Failed to insert %s@%s: %v", msg.Chat, msg.JID, err)
	}
}
