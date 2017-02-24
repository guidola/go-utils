package database


import (
	"time"
	"os"
	"github.com/labstack/gommon/log"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"bytes"
	"sync"
	"errors"
	"net/url"
)


const NeoMaxPoolSize = 100 //this value is greatly lower than a real one, see https://jaksprats.wordpress.com/2010/09/22/12/
const NeoPoolSize = 50
const NeoURI="NEO_URI"

type CypherQuery struct {

	Query string
	Params map[string] interface{}

}

func (cq CypherQuery) GetCopy() CypherQuery {

	return cq
}

func (cq CypherQuery) GetCopyWithParams(params map[string]interface{}) CypherQuery {
	cq.Params = params
	return cq
}


// ****************************************************************

type NeoConfig struct {

	Protocol, Url, Username, Password string
	Size, Max_size int  //max_size is not yet usable since pool abstraction on bolt driver does not allow to create a two range pool
	Requires_auth bool

}

/**
	Formats the configuration into the url that has to be used to connecto to the noe4j database using the http api
 */
func (nc NeoConfig) getURI() string {
	var buffer bytes.Buffer
	buffer.WriteString(nc.Protocol)
	buffer.WriteString("://")
	if nc.Requires_auth {
		buffer.WriteString(nc.Username)
		buffer.WriteString(":")
		buffer.WriteString(url.QueryEscape(nc.Password))
		buffer.WriteString("@")
	}
	buffer.WriteString(nc.Url)
	return buffer.String()
}

var DefaultNeoConfig = NeoConfig{
	Protocol: "bolt",
	Url: os.Getenv(NeoURI),
	Username: "neo4j",
	Password: "guillermo95dobao",
	Size: NeoPoolSize,
	Max_size: NeoMaxPoolSize,
	Requires_auth: true,
}

var AuthNeoConfig = NeoConfig{
	Protocol: "bolt",
	Url: os.Getenv(NeoURI),
	Username: "neo4j",
	Password: "=AW:=Gf7SS;N+*K-:m|OJ=Vosn.lBt+%f.pKLl!%D|4^I%X2Z^7_1-Y|fkkRMk_ufc*JOo|J1hK4Ycs^SZSqQVeOd~Iqg;07%t1~%6.Mm~_s.e9O13l.!%.L!IamV0|y",
	Size: NeoPoolSize,
	Max_size: NeoMaxPoolSize,
	Requires_auth: true,
}

// ****************************************************************

type Neo struct{

	uri string
	aux_driver bolt.Driver  //this driver is used to give extra connections when the pool is empty till a max
	pool bolt.DriverPool
	pool_empty bool
	extra int           //maps the extra capacity we're faking for the pool
	connected bool

}

var neo_instance Neo


/**
	Returns a pointer to the instance stored in redis private variable
 */
func GetNeoInstance() *Neo{
	return &neo_instance
}

// *******************************PATCH AREA FOR POOL****************************************

type BoltConn struct {
	Client bolt.Conn
	IsPooled bool
}

func (bc * BoltConn) Close() error {
	return bc.Client.Close()
}

type connection struct {
	conn bolt.Conn
	err error
}

func newConn(ch chan connection, pool bolt.DriverPool) {

	conn, err := pool.OpenPool()
	ch <- connection{conn: conn, err: err}

}

func newConnTimeout(ch chan bool) {
	time.Sleep(1 * time.Millisecond)  //esperem un milisegon
	ch <- true
}

var mutex = sync.Mutex{}
var timeout = make(chan bool, 1)
var conns = make(chan connection, 1)

/**
	Masks the newPool() method from bolt driver to avoid getting blocked if there are no connections left on the pool.
	In that case return a new single connection.

	IMPORTANT:  consider this has no limit so you could eventually exceed the max value on the db and weird things
				could happen.
 */
func getConnection(n *Neo) (BoltConn, error) {

	mutex.Lock()
	defer mutex.Unlock()

	if n.pool_empty {
		go func(){timeout<-true}()
	} else {
		go newConn(conns, n.pool)
		go newConnTimeout(timeout)
	}

	select {
	case new_conn  := <-conns:
		return BoltConn{Client:new_conn.conn, IsPooled: true}, new_conn.err

	case <-timeout:
		n.pool_empty = true

		//create new conn and return
		if n.extra > 0 {
			n.extra = n.extra - 1
			conn, err := n.aux_driver.OpenNeo(n.uri)
			return BoltConn{Client:conn, IsPooled: false}, err
		}

		//si ja hem creat el maxim aleshores si que bloquejem.
		conn, err := n.pool.OpenPool()
		return BoltConn{Client:conn, IsPooled: true}, err
	}
}

// *********************************************************************************************************

func (n * Neo) CreateWithConfig(config NeoConfig) error {

	return n.Create(config.getURI(), config.Size, config.Max_size)
}


/**
	Creates a neo pool using the configuration provided.
 */
func (n *Neo) Create(conn_url string, size int, max_size int) error {
	log.Info(conn_url)
	//create pool
	_pool, err := bolt.NewDriverPool(conn_url, size)
	if err != nil {
		log.Fatalf("error %s creating connection pool against neo4j instance at %s with a size of %d",
			err.Error(), conn_url, size)
		return err
	}

	//create backup driver for extra connections
	n.uri = conn_url
	n.pool_empty = false
	n.aux_driver = bolt.NewDriver()
	n.pool = _pool
	n.connected = true
	n.extra = max_size - size

	return nil
}


/**
	Executes the given command with the given parameters automatically requesting for a connection against the pool.
	returns the result and err as return parameters
 */
func (n * Neo) Execute(query CypherQuery) (bolt.Result, error) {

	//retrieve connection
	conn, err := n.Get()
	if err != nil {
		return nil, err
	}
	defer n.Put(conn)

	result, err := conn.Client.ExecNeo(query.Query, query.Params)

	if err != nil {
		return nil, err
	}

	return result, nil

}


/**
	Executes the given query with the given parameters automatically requesting for a connection against the pool.
	returns the result and err as return parameters
 */
func (n * Neo) Query(query CypherQuery) (bolt.Rows, BoltConn, error) {

	//retrieve connection
	conn, err := n.Get()
	if err != nil {
		return nil, BoltConn{}, err
	}

	result, err := conn.Client.QueryNeo(query.Query, query.Params)

	if err != nil {
		return nil, conn, err
	}

	return result, conn, nil

}


/**
	Masks the OpenPool Function to retrieve a new connection from the pool.
 */
func (n * Neo) Get() (BoltConn, error){

	if  n.connected {
		return getConnection(n)
	} else {
		return BoltConn{}, errors.New("The Neo4j pool is closed")
	}

}


/**
	Masks the pool put function for direct access from the neo structure
 */
func (n * Neo) Put(c BoltConn) error {

	//if it is an extra connection just inc the counter of available extra
	if !c.IsPooled {
		n.extra = n.extra + 1
	}
	// always execute the close and return the result
	return c.Client.Close()
}


/**
	Masks the pool Empty function for direct access from the Redis structure
 */
func (n * Neo) Destroy() {

	//the library does not support closing of pool connections so we do it for em
	timeout2 := make(chan bool, 1)
	conns := make(chan bool, 1)
	done := false
	for !done {

		go func() {
			conn, _ := n.pool.OpenPool()
			// set a really low timeout value to time the connection out. since there's no way to actually close a
			// pooled connection through the bolt.Conn exposed methods.
			conn.SetTimeout(time.Millisecond)
			conns <- true
		}()

		go func() {
			time.Sleep(50 * time.Millisecond)
			timeout2 <- true
		}()

		select {
		case <- conns:
			//OK
		case <- timeout2:
			done = true
		}
	}

	n.connected = false

	// this is kind of fake since if there are connections that have not been returned to the pool
	// if are returned once this method is executed it will contain open connections but be
	// marked as non connected anyway.
}

