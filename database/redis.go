package database

import(
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/labstack/gommon/log"
	"errors"
	"time"
)

/*
	Redis serves as a singleton like abstraction for global use of a redis connection pool inside a project
	that uses a single instance of a redis pool. The pool itself cannot be directly accessed but has to be
	interacted with using the methods described below
 */
type Redis struct{

	pool *pool.Pool
	connected bool

}

var redis_instance Redis


/*
	Returns the redis instance either it has been initialized or not.
 */
func GetRedisInstance() *Redis{
	return &redis_instance
}

/*

	Create Creates a redis pool pointing to the given address and of the given size using the secret specified as a
	parameter to authenticate against the redis database.

	The pool is created even if there is an error creating the connections. So if later on the db begins to be
	reachable new connections can be created on the go and stored inside the pool afterwards until the limit set by size.

 */
func (r *Redis) CreateWithCredentials(protocol string, remote_endpoint string, secret string, size int) error {

	auth_routine := func(network, addr string) (*redis.Client, error) {
		client, err := redis.DialTimeout(network, addr, time.Second * 60)
		if err != nil {
			return nil, err
		}
		if err = client.Cmd("AUTH", secret).Err; err != nil {
			client.Close()
			return nil, err
		}
		return client, nil
	}

	pool_instance, err := pool.NewCustom(protocol, remote_endpoint, size, auth_routine)

	if err != nil {
		log.Errorf("error %s creating connection pool against redis instance at %s://%s with a size of %d",
			err.Error(), protocol, remote_endpoint, size)
		return err
	}

	r.pool = pool_instance
	r.connected = true
	return nil
}


/*
	Create Creates a redis pool pointing to the given address and of the given size.

	The pool is created even if there is an error creating the connections. So if later on the db begins to be
	reachable new connections can be created on the go and stored inside the pool afterwards until the limit set by size.
 */
func (r *Redis) Create(protocol string, remote_endpoint string, size int) error {


	pool_instance, err := pool.New(protocol, remote_endpoint, size)
	if err != nil {
		log.Errorf("error %s creating connection pool against redis instance at %s://%s with a size of %d",
			err.Error(), protocol, remote_endpoint, size)
		return err
	}

	r.pool = pool_instance
	r.connected = true
	return nil
}


/*
	Executes the given command with the given parameters automatically requesting for a connection against the pool
	and returning it afterwards.

	Returns the response and response.Err as return parameters
 */
func (r* Redis) Execute(command string, args ...interface{}) (*redis.Resp, error) {

	resp := r.pool.Cmd(command, args)
	return resp, resp.Err
}


/*
	Masks the pool get function for direct access from the Redis structure
 */
func (r* Redis) Get() (*redis.Client, error){
	if r.connected {
		return r.pool.Get()
	}

	return nil, errors.New("Connection pool has been destroyed. Initialize it again before requesting " +
		"more connections")
}


/*
	Masks the pool put function for direct access from the Redis structure
 */
func (r* Redis) Put(c *redis.Client){
	if r.connected {
		r.pool.Put(c)
		return
	}

	c.Close()
}


/*
	Masks the pool Empty function for direct access from the Redis structure. It also invalidates the redis struct
	so no further connections can be requested. And returned connections are closed instead.
 */
func (r* Redis) Destroy() {
	r.pool.Empty()
	r.connected = false // this is kind of fake since if there are connections that have not been returned to the pool
						// if are returned once this method is executed it will contain open connections but be
						// marked as non connected anyway.
}