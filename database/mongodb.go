package database

import (

	"gopkg.in/mgo.v2"
	"time"
	"github.com/labstack/gommon/log"
	"strings"
)


/*
	Mongo serves as a singleton like abstraction for global use of a mongodb master session inside a project
	that uses a single instance of a master session. The session itself cannot be directly accessed but has to be
	interacted with using the methods described below
 */
type Mongo struct{

	master_session *mgo.Session

}

var (
	mongo_instance Mongo

	default_dial_info = &mgo.DialInfo{

		Timeout: 60 * time.Second,
		FailFast: true,
		Database: "test",
		Source: "admin",
		PoolLimit: 100, //this value is not realistic but enough for developing. hard limit on mongo is 24000 by default
	}
)


/*
	GetMongoInstance Returns a reference to the global instance of MongoDB which provides functionalities to establish new sessions
	the database
 */
func GetMongoInstance() *Mongo{
	return &mongo_instance
}


func configureDefaultSession(session *mgo.Session) {
	mongo_instance.master_session = session
	mongo_instance.master_session.SetMode(mgo.Monotonic, true)
	session.EnsureSafe(&mgo.Safe{W: 1, FSync: true})    	// sets to 1 the number of servers that gotta flush
								// changes to disc to consider an operation satisfactory
}


/*
	DefaultConfigWithHosts return the default configuration specified at DefaultConfig but already initializes the
	hosts to connect to.

	Should we then expose the DefaultConfig constructor ?
 */
func (m Mongo) DefaultConfigWithHosts (hosts []string) *mgo.DialInfo {
	conf := m.DefaultConfig()
	conf.Addrs = hosts
	return conf
}

/*
	DefaultConfig Returns a mgo.DialInfo struct with some prefilled options. See https://godoc.org/gopkg.in/mgo.v2#DialInfo
	for further information on how mgo.DialInfo works.

	The default parameters set in this default configuration are:

		Timeout: 60 seconds,
		FailFast:  true,
		Database: test,
		Source: admin,
		PoolLimit: 100

	Its important to note that PoolLimit is set to a low value for development and basic testing purposes. For
	production environments this value should be set to the maximum of connections you want an instance of your
	server to establish against the given mongodb cluster

	Further configuration can be done using this as a baselina or even modification of the default parameters
	can occur.

 */
func (m Mongo) DefaultConfig () *mgo.DialInfo {

	new_config := new(mgo.DialInfo)
	*new_config = *default_dial_info
	return new_config
}


/*
	Create Creates a master session against the cluster identified by the provided hosts configured by default with monotonic
	mode and safety parameters that consider a write OK when 1 hots of the replica set has flushed the changes to disc.

	This Create function does not take any config and establishes the master session using the default configuration
	parameters for mgo sessions. For more advanced configuration use CreateWitConfig function.
 */
func (m *Mongo) Create(hosts []string) error {

	session, err := mgo.Dial(strings.Join(hosts, ","))

	if err != nil {
		log.Fatalf("Default dial to mongodb cluster located at %s failed with error %s...",
			hosts,
			err.Error())

		return err
	}

	configureDefaultSession(session)

	return nil
}


/*
	CreateWithConfig Creates a master session against the cluster identified by the provided hosts configured by default with monotonic
	mode and safety parameters that consider a write OK when 1 hots of the replica set has flushed the changes to disc

	This Create function does not take any config and establishes the master session using the default configuration
	parameters for mgo sessions. For more advanced configuration use CreateWitConfig function.

	The configuration struct follows the conventions explained at https://godoc.org/gopkg.in/mgo.v2#DialInfo.
	For a default configuration at which you only need to specify the hosts to connect to see DefaultConfig
 */
func (m *Mongo) CreateWithConfig(dial_info *mgo.DialInfo) error {

	session, err := mgo.DialWithInfo(dial_info)

	if err != nil {
		log.Fatalf("Dial with info to mongodb cluster located at %s failed with error %s...",
			dial_info.Addrs,
			err.Error())

		return err
	}

	configureDefaultSession(session)

	return nil
}


/*
	Returns a copy of the master session running on a new socket.
 */
func (m *Mongo) GetCopy() *mgo.Session{

	return mongo_instance.master_session.Copy()
}


//TODO Consider creating different functions for retrieving sessions with different configurations such as work mode

/*
	Returns a copy of the master session running on a new socket with batching enabled by default for the batch size
	provided on the call to this function
 */
func (m *Mongo) GetStreamedCopy(batch_size int) *mgo.Session{

 	new_session := mongo_instance.master_session.Copy()
	new_session.SetBatch(batch_size)

	return new_session
}


/*
	Destroy closes the main Session therefore not allowing for further connections to be established.
	TODO check wether this invalidates all sessions or just closes main session. Also check if pooled connections are disconnected
 */
func (m *Mongo) Destroy() {

	//i guess i just have to close the main session.
	mongo_instance.master_session.Close()

}