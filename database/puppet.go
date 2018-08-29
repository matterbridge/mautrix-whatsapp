// mautrix-whatsapp - A Matrix-WhatsApp puppeting bridge.
// Copyright (C) 2018 Tulir Asokan
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
	"database/sql"

	log "maunium.net/go/maulogger"
	"maunium.net/go/mautrix-whatsapp/types"
)

type PuppetQuery struct {
	db  *Database
	log log.Logger
}

func (pq *PuppetQuery) CreateTable() error {
	_, err := pq.db.Exec(`CREATE TABLE IF NOT EXISTS puppet (
		jid         VARCHAR(25) PRIMARY KEY,
		displayname VARCHAR(255),
		avatar      VARCHAR(255)
	)`)
	return err
}

func (pq *PuppetQuery) New() *Puppet {
	return &Puppet{
		db:  pq.db,
		log: pq.log,
	}
}

func (pq *PuppetQuery) GetAll() (puppets []*Puppet) {
	rows, err := pq.db.Query("SELECT * FROM puppet")
	if err != nil || rows == nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		puppets = append(puppets, pq.New().Scan(rows))
	}
	return
}

func (pq *PuppetQuery) Get(jid types.WhatsAppID) *Puppet {
	row := pq.db.QueryRow("SELECT * FROM puppet WHERE jid=?", jid)
	if row == nil {
		return nil
	}
	return pq.New().Scan(row)
}

type Puppet struct {
	db  *Database
	log log.Logger

	JID         types.WhatsAppID
	Displayname string
	Avatar      string
}

func (puppet *Puppet) Scan(row Scannable) *Puppet {
	var displayname, avatar sql.NullString
	err := row.Scan(&puppet.JID, &displayname, &avatar)
	if err != nil {
		if err != sql.ErrNoRows {
			puppet.log.Errorln("Database scan failed:", err)
		}
		return nil
	}
	puppet.Displayname = displayname.String
	puppet.Avatar = avatar.String
	return puppet
}

func (puppet *Puppet) Insert() error {
	_, err := puppet.db.Exec("INSERT INTO puppet VALUES (?, ?, ?)",
		puppet.JID, puppet.Displayname, puppet.Avatar)
	if err != nil {
		puppet.log.Errorfln("Failed to insert %s: %v", puppet.JID, err)
	}
	return err
}

func (puppet *Puppet) Update() error {
	_, err := puppet.db.Exec("UPDATE puppet SET displayname=?, avatar=? WHERE jid=?",
		puppet.Displayname, puppet.Avatar, puppet.JID)
	if err != nil {
		puppet.log.Errorfln("Failed to update %s->%s: %v", puppet.JID, err)
	}
	return err
}
