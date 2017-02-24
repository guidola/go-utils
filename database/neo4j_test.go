package database

import (
	"testing"
	"io"
	"time"
)

func TestNeo4JLifeCycle(t * testing.T){
	
	var conf = DefaultNeoConfig
	conf.Size = 2
	conf.Max_size = 3
	
	neo_instance_test :=  GetNeoInstance()
	
	//connect to the database, expect it connects with no error
	err := neo_instance_test.CreateWithConfig(conf)
	if err != nil {
		t.Errorf("Expected to get no Error and got %s", err.Error())
		t.FailNow()
	}
	
	//insert a test value it should not get an error
	cq := CypherQuery{
		Query: `CREATE (a:Test {name:'Arthur', title:'King'})`,
	}
	res, err := neo_instance_test.Execute(cq)
	if err != nil {
		t.Errorf("Expected to get no Error and got %s", err.Error())
		t.FailNow()
	}

	//check there is one affected row to check that response is correctly processed
	if ra, _ := res.RowsAffected(); ra != 1 {
		t.Errorf("Expected to get one affected row and got %d", ra)
	}

	//check if previously inserted value exists
	cq = CypherQuery{
		Query: `MATCH (a:Test) RETURN a.name, a.title`,
	}
	q_res, conn, err := neo_instance_test.Query(cq)
	if err != nil {
		t.Errorf("Expected to get no Error and got %s", err.Error())
		t.FailNow()
	}

	if q_res == nil {
		t.Error("Expected to get Rows entity and got nil")
		t.FailNow()
	}

	if f_row, _, err := q_res.NextNeo(); err == io.EOF {
		t.Error("Got EOF when expected one result")
	} else if err != nil {
		t.Errorf("Expected no error and got %s", err.Error())
	} else if f_row[0].(string) != "Arthur" || f_row[1].(string) != "King" {
		t.Errorf("Expected to get a node with properties name:'Arthur' and title:'King' and got %s %s instead",
			f_row[0].(string),
			f_row[1].(string),
		)
	}

	conn.Close()

	// remove inserted node
	res, err = neo_instance_test.Execute(CypherQuery{Query:`MATCH (a:Test) DELETE a`})

	if err != nil {
		t.Error("something went wrong deleting the test node")
	}

	conn, err = neo_instance_test.Get()

	err = neo_instance.Put(conn)

	// test that we are able to get a connection when we have taken an returned a connection

	//works until here. next stuff may even not be implemented this way so im letting it crash for now.

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()

	getdone := make(chan bool, 1)
	go func() {
		_, err = neo_instance_test.Get()
		if err != nil {
			t.Errorf("got %s error when expecting new connection because of asking connection a pool wiht remaining " +
				"connections", err.Error())
		}
		getdone <- true
	}()

	select {
	case <- timeout:
		t.Error("Could not retrieve a connection when requesting one to non-empty pool. Got deadlocked")
	case <- getdone:
		//OK
	}


	// check that if we get all the connections from the pool a new one is returned to us
	var connections [2]BoltConn
	for i := 0 ; i < 2 ; i++{
		connections[i], _ = neo_instance_test.Get()
	}
	
	// test that we are able to get a new connection event that we have emptied the pool
	extra_con1, err := neo_instance_test.Get()
	if(err != nil){
		t.Errorf("got %s error when expecting new connection because of asking connection to an empty pool", err.Error())
	}

	// test that we are able to return connection 1 to the pool but
	// return the connections to the pool
	for i := 0 ; i < 2 ; i++{
		neo_instance_test.Put(connections[i])
	}
	if err = neo_instance_test.Put(extra_con1); err != nil {
		t.Error("Expected the connection to be returned to the pool but it was already full")
	}

	//destroy the pool freeing all the connections on the pool, if some connection has not been returned it will not
	// be timed out.
	neo_instance_test.Destroy()

	if _, err = neo_instance.Get(); err == nil {
		t.Error("Could retrieve a connection when expected not to since i closed the pool")
	}
	
}