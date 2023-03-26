package db

import (
	"github.com/pwera/app/domain"
	mgo "gopkg.in/mgo.v2"
	"io"
	"log"
	"net"
	"time"
)

type Connector struct {
	db     *mgo.Session
	conn   net.Conn
	reader io.ReadCloser
}

func NewConnector() *Connector {
	log.Println("dialing mongodb: mongo")
	ddb, err := mgo.Dial("mongo:27017")
	if err != nil {
		log.Fatalf("MongoDb dialing problem %v\n", err)
	}
	return &Connector{db: ddb}
}

func (c *Connector) CloseDb() {
	c.db.Close()
	log.Println("Connection closed")
}

func (c *Connector) CloseConn() {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.reader != nil {
		c.reader.Close()
	}
}

func (c *Connector) LoadOptions() ([]string, error) {
	var options []string
	iter := c.RetrivePollData().Find(nil).Iter()
	var p domain.Poll
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

func (c *Connector) RetrivePollData() *mgo.Collection {
	return c.db.DB("ballots").C("polls")
}

func (c *Connector) Dial(netw, addr string) (net.Conn, error) {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	c.conn = netc
	return netc, nil
}
func (c *Connector) SessionCopy() *mgo.Session {
	return c.db.Copy()
}
