package main

import (
	"errors"
	"fmt"
	"github.com/gocraft/dbr/v2"
	_ "github.com/lib/pq"
	state_machine "state-machine"
)

func main() {

	// database
	conn, err := dbr.Open("postgres", "host=localhost auth=postgres password=password dbname=yonderland sslmode=disable", nil)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()
	sessionBD := conn.NewSession(nil)
	createTableAndToTest(sessionBD, false)

	// database tx
	tx, err := sessionBD.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	// state machine 1
	sm1 := state_machine.NewStateMachine()
	err = sm1.Load("examples/example-order-state-machine.json")
	if err != nil {
		fmt.Println("Error parsing state machine:", err)
		return
	}
	sm1.AddExecuteFunction(UpdateStatusOrder)
	sm1.AddCheckFunction("auth", Auth)
	sm1.AddOnErrorFunction("trigger_error", TriggerError)

	// state machine 2
	sm2 := state_machine.NewStateMachine()
	err = sm2.Load("examples/example-item-order-state-machine.json")
	if err != nil {
		fmt.Println("Error parsing state machine:", err)
		return
	}
	sm2.AddExecuteFunction(UpdateStatusOrderItem)
	sm1.AddStateMachineToTrigger(sm2)

	// state machine
	obj := ObjStateMachine{
		Tx:            tx,
		Authorization: []string{"AUTH_1"},
		IdOrder:       1,
		IdProduct:     1,
	}
	err = sm1.ProcessTransition("being-processed", "ready-for-pickup", obj)
	if err := tx.Commit(); err != nil {
		fmt.Println("Commit error")
	}

	fmt.Println("state machine change successful the state")

}

type ObjStateMachine struct {
	Authorization []string `json:"authorization"`
	Tx            *dbr.Tx  `json:"tx"`
	IdProduct     int      `json:"id_product"`
	IdOrder       int      `json:"id_order"`
}

func UpdateStatusOrder(currentState, nextState string, obj any) (bool, error) {
	smObj := obj.(ObjStateMachine)

	var fkStatusNexState int

	_, err := smObj.Tx.Select("id_status").
		From("test.status").
		Where("key = ?", nextState).
		Load(&fkStatusNexState)
	if err != nil {
		return false, err
	}

	_, err = smObj.Tx.Update("test.order").
		Set("fk_status", fkStatusNexState).
		Where("id_order = ?", smObj.IdOrder).
		Exec()

	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateStatusOrderItem(currentState, nextState string, obj any) (bool, error) {
	smObj := obj.(ObjStateMachine)

	var fkStatusNexState int

	_, err := smObj.Tx.Select("id_status").
		From("test.status").
		Where("key = ?", nextState).
		Load(&fkStatusNexState)
	if err != nil {
		return false, err
	}

	_, err = smObj.Tx.Update("test.product").
		Set("fk_status", fkStatusNexState).
		Where("id_product = ?", smObj.IdProduct).
		Exec()

	if err != nil {
		return false, err
	}

	return true, nil
}

func Auth(obj any, input ...string) (bool, error) {
	smObj := obj.(ObjStateMachine)
	for _, in := range input {
		for _, userAuth := range smObj.Authorization {
			if userAuth == in {
				return true, nil
			}
		}
	}

	return false, errors.New("don't have authorization")
}

func TriggerError(obj any, input ...string) (bool, error) {
	return true, errors.New("error in x")
}

func createTableAndToTest(session *dbr.Session, migration bool) {
	var err error
	if migration {
		// Create Status table
		_, err = session.Exec(
			`CREATE TABLE IF NOT EXISTS test.status (
        id_status SERIAL PRIMARY KEY,
        key VARCHAR(50))`)
		if err != nil {
			panic(err.Error())
		}

		// Create Order table
		_, err = session.Exec(
			`CREATE TABLE IF NOT EXISTS test.order (
        id_order SERIAL PRIMARY KEY,
        fk_status INT REFERENCES test.status(id_status),
        description VARCHAR(255))`)
		if err != nil {
			panic(err.Error())
		}

		// Create Product table
		_, err = session.Exec(
			`CREATE TABLE IF NOT EXISTS test.product (
        id_product SERIAL PRIMARY KEY,
        fk_status INT REFERENCES test.status(id_status),
    	fk_order INT REFERENCES test.order(id_order),
        description VARCHAR(255))`)
		if err != nil {
			panic(err.Error())
		}

		stats := []string{"being-processed", "ready-for-pickup", "ready-for-shipment", "cancel"}

		for _, s := range stats {
			_, err = session.InsertInto("test.status").
				Columns("key").Values(s).
				Exec()
			if err != nil {
				return
			}
		}

		_, err = session.InsertInto("test.order").
			Columns([]string{
				"fk_status",
				"description",
			}...).
			Values(
				1,
				"test1",
			).Exec()
		if err != nil {
			return
		}

		_, err = session.InsertInto("test.product").
			Columns([]string{
				"fk_status",
				"description",
				"fk_order",
			}...).
			Values(
				1,
				"test1",
				1,
			).Exec()
		if err != nil {
			return
		}

		fmt.Println("Tables created successfully!")
	}
}
