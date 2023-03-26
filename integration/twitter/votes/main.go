package main

import (
	"flag"
	"github.com/nsqio/go-nsq"
	"github.com/pwera/app/db"
	"github.com/pwera/app/twitter"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var host string
	flag.StringVar(&host, "host", "localhost", "host")
	flag.Parse()
	twitter := twitter.Twitter{Connector: db.NewConnector()}
	defer twitter.Connector.CloseDb()

	var stoplock sync.Mutex
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping")
		stopChan <- struct{}{}
		twitter.Connector.CloseConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes, host)
	twitterStoppedChan := twitter.StartTwitterStream(stopChan, votes)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			twitter.Connector.CloseConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				return
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan

}

func publishVotes(votes <-chan string, host string) <-chan struct{} {
	stopchan := make(chan struct{}, 1)
	pub, _ := nsq.NewProducer(host+":4150", nsq.NewConfig())
	go func() {
		for vote := range votes {
			pub.Publish("votes", []byte(vote))
		}
		log.Println("Producer stopping")
		pub.Stop()
		log.Println("Producer stopped")
		stopchan <- struct{}{}
	}()
	return stopchan
}
