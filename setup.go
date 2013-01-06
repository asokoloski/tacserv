package main

import (
	"code.google.com/p/gosqlite/sqlite"
)

// exits program on failure
func setup(db *sqlite.Conn) error {
	statements := []string{
		"create table if not exists tool (node_id int, name text, status int)",
		"create table if not exists permission (node_id int, timestamp datetime default (CURRENT_TIMESTAMP), card_id text, granter_card_id text, is_maintainer boolean default 0)",
		"create unique index if not exists permission_node_card on permission (node_id, card_id)",
		"create table if not exists tool_usage (node_id int, timestamp datetime default (CURRENT_TIMESTAMP), status int, card_id text)",
		"create table if not exists case_alert (node_id int, timestamp datetime default (CURRENT_TIMESTAMP), status int)",
	}

	for _, stmt := range statements {
		err := db.Exec(stmt)
		if err != nil {
			log("error executing:", stmt)
			return err
		}
	}
	return nil
}
