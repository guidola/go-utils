package database

import (
	"testing"
	"github.com/mediocregopher/radix.v2/redis"
	"os"
)

func TestRedisLifeCycle(t * testing.T){

	var redisURI = os.Getenv("REDIS_URI")

	redis_instance_test :=  GetRedisInstance()

	//connect to the database, expect it connects with no error
	err := redis_instance_test.Create("tcp", redisURI, 2)
	if err != nil {
		t.Errorf("Expected to get no Error and got %s", err.Error())
		t.FailNow()
	}

	//insert a test value it should not get an error
	resp, err := redis_instance_test.Execute("SET", "foo", "bar", "EX", 2)
	if err != nil {
		t.Errorf("Expected to get no Error and got %s", err.Error())
		t.FailNow()
	}

	//check if previously inserted value exists
	resp, err = redis_instance_test.Execute("EXISTS", "foo")
	if err != nil {
		t.Errorf("Expected to get no Error and got %s", err.Error())
		t.FailNow()
	}
	if val, _ := resp.Int(); val != 1 {
		t.Error("Expected for 'foo' key to exist and got that it does not")
	}

	//check that if we get all the connections from the pool we cannot retrieve more
	var connections [2]*redis.Client
	for i := 0 ; i < 2 ; i++{
		connections[i], _ = redis_instance_test.Get()
	}

	//test that we cannnot get a new connection
	_, err = redis_instance_test.Get()
	if err != nil {
		t.Errorf("got %s error when expecting new connection because of asking connection to an empty pool", err.Error())
	}

	// return the connections to the pool
	for i := 0 ; i < 2 ; i++{
		redis_instance_test.Put(connections[i])
	}

	//destroy the pool freeing all the connections on the pool, if some connection has not been returned it will not
	// be freed.
	redis_instance_test.Destroy()


}