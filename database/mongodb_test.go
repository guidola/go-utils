package database

import (
	"testing"
	"os"
)

func TestMongoDBLifecycle(t * testing.T){

	mongoHosts := []string{os.Getenv("MONGO_URI")}

	mongo := GetMongoInstance()

	err := mongo.Create(mongoHosts)
	if err != nil {
		t.Errorf("There was an error while connecting to the mongo instance: %s", err.Error())
		t.FailNow()
	}

	//to check strictly the functionalities I should check that the configurations made against the master session
	//are actually working not just test reachability for now we're settling for that since the underlying library is suposed to work

	if err := mongo.master_session.Ping(); err != nil{
		t.Errorf("Expected for the mongodb cluster to be reachable and it is not reachable instead with error %s",
		err.Error())
	}

	mongo.Destroy()

	defer func(){
		recover()
	}()
	mongo.master_session.Ping()
	t.Error("Got to this statement when shouldnt because a panic should have been triggered")

}

func TestMongoDBLifecycleWithConfig(t * testing.T){

	mongoHosts := []string{os.Getenv("MONGO_URI")}

	mongo := GetMongoInstance()

	dial_info := mongo.DefaultConfigWithHosts(mongoHosts)

	err := mongo.CreateWithConfig(dial_info)
	if err != nil {
		t.Errorf("There was an error while connecting to the mongo instance: %s", err.Error())
		t.FailNow()
	}

	//to check strictly the functionalities I should check that the configurations made against the master session
	//are actually working not just test reachability for now we're settling for that since the underlying library is suposed to work

	if err := mongo.master_session.Ping(); err != nil{
		t.Errorf("Expected for the mongodb cluster to be reachable and it is not reachable instead with error %s",
			err.Error())
	}

	mongo.Destroy()

	defer func(){
		recover()
	}()
	mongo.master_session.Ping()
	t.Error("Got to this statement when shouldnt because a panic should have been triggered")

}