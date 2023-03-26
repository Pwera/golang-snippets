package main

import (
	"flag"
	"fmt"
	nsq "github.com/nsqio/go-nsq"
	"github.com/pwera/app/db"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type data map[string]int

const updateDuration = 1 * time.Second

func main() {
	var host string
	flag.StringVar(&host, "host", "mongo", "host")
	flag.Parse()
	connector := db.NewConnector()
	fmt.Print("Couter")
	options, err := connector.LoadOptions()
	pollData := connector.RetrivePollData()
	_ = pollData
	defer connector.CloseDb()
	if err != nil {
		log.Fatalf("Couldn't load option %v", err)
	}
	_ = options
	var counts data
	var countsLock sync.Mutex
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		log.Fatalf("Nsq consumer error %v\n", err)
	}
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(data)
			vote := string(m.Body)
			counts[vote]++
		}
		return nil
	}))
	if err := q.ConnectToNSQLookupd(host + ":4161"); err != nil {
		log.Fatalf("ConnectToNSQLookupd problem %v\n", err)
		return
	}

	ticker := time.NewTicker(updateDuration)
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case <-ticker.C:
			doCounter(&countsLock, &counts)
			q.Stop()
		case <-q.StopChan:
			return
		}
	}

	time.Sleep(50 * time.Second)
}

func doCounter(lock *sync.Mutex, counts *data /*, pollData *mgo.Collection*/) {
	lock.Lock()
	defer lock.Unlock()
	if len(*counts) == 0 {
		log.Println("No votes")
		return
	}
	ok := true
	for option, count := range *counts {
		_ = option
		_ = count
		//sel := bson.M{"options": bson.M{"$in": []string{option}}}
		//up := bson.M{"$inc": bson.M{"results." + option: count}}
		//if _, err := pollData.UpdateAll(sel, up); err != nil {
		//	log.Printf("Database update action problem %v\n", err)
		//	ok = false
		//}
	}
	if ok {
		*counts = nil
	}
}
