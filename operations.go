package main

import (
	"errors"
)

var TimeFormat = "2006-01-02T15:03:04.000"

func getPermissions(nodeId int, cardId string) (int, error) {
	stmt, err := db.Prepare("select is_maintainer from permission where node_id = ? and card_id = ?")
	if err != nil {
		return 0, err
	}
	err = stmt.Exec(nodeId, cardId)
	if err != nil {
		return 0, err
	}
	isMaintainer := false
	allowed := false
	for stmt.Next() {
		allowed = true
		stmt.Scan(isMaintainer)
	}
	if allowed {
		if isMaintainer {
			return 2, nil
		}
		return 1, nil
	}
	return 0, nil
}

func setPermissions(nodeId int, cardId string, granterCardId string, level int) (int, error) {
	var cmd string
	var params []interface{}
	switch level {
	case 0:
		cmd = "delete from permission where node_id = ? and card_id = ?"
		params = []interface{}{nodeId, cardId}
	case 1:
		cmd = "insert or replace into permission (node_id, card_id, granter_card_id) values (?, ?, ?)"
		params = []interface{}{nodeId, cardId, granterCardId}
	case 2:
		cmd = "insert or replace into permission (node_id, card_id, granter_card_id, is_maintainer) values (?, ?, ?, 1)"
		params = []interface{}{nodeId, cardId, granterCardId}
	default:
		return 0, errors.New("Invalid permission level")
	}
	err := db.Exec(cmd, params...)
	if err != nil {
		log(err)
		return 0, err
	}
	return 1, nil
}

func getNextCard(nodeId int, lastCardId string) (string, error) {
	log(nodeId, lastCardId)
	stmt, err := db.Prepare("select card_id from permission where node_id = ? and card_id > ? limit 1")
	if err != nil {
		log(err)
		return "", err
	}
	err = stmt.Exec(nodeId, lastCardId)
	if err != nil {
		log(err)
		return "", err
	}
	newCardId := new(string)
	for stmt.Next() {
		stmt.Scan(newCardId)
	}
	return *newCardId, nil
}
