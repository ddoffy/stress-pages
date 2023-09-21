package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

var (
	total         int = 0
	totalRequests int = 0
	// current connections to the host (global) (int)
	currentConnections chan int
	disconnections     chan int
	currentRequests    chan int
	// time in seconds
	seconds    int = 0
	ttl        int = 300
	domainName string
	host       string
)

func display() {
	for {
		select {
		case <-currentConnections:
			total++
			fmt.Printf("Current Connections: %d - Current Requests: %d - The secons: %d\n", total, totalRequests, seconds)
		case <-disconnections:
			total--
			fmt.Printf("Current Connections: %d - Current Requests: %d - The secons: %d\n", total, totalRequests, seconds)
		case <-currentRequests:
			totalRequests++
			fmt.Printf("Current Connections: %d - Current Requests: %d - The secons: %d\n", total, totalRequests, seconds)
		}
	}
}

func handler() {
	currentConnections = make(chan int, 10)
	disconnections = make(chan int, 10)
	currentRequests = make(chan int, 10)
	go display()
	numberOfTask := 10
	url := "https://www.google.com"

	args := os.Args[1:]

	if len(args) > 0 {
		url = args[0]
	}

	if len(args) > 1 {
		num, err := strconv.Atoi(args[1])

		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		numberOfTask = num
	}

	splitUrl := strings.Split(url, "/")
	domainName = splitUrl[2]
	host = splitUrl[0] + "//" + splitUrl[2]

	fmt.Println("Host:", host)
	fmt.Println("Domain Name:", domainName)

	var wg sync.WaitGroup
	wg.Add(numberOfTask)
	for i := 0; i < numberOfTask; i++ {
		go func(link string) {
			defer func() {
				disconnections <- 1
				wg.Done()
			}()
			currentConnections <- 1
			collect(link)
		}(url)
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
}

func main() {
	Lambda.Start(handler)
}

func collect(url string) {
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		nUrl := e.Attr("href")
		if strings.HasPrefix(nUrl, host) {
			e.Request.Visit(nUrl)
		} else if strings.Contains(nUrl, domainName) {
			e.Request.Visit(nUrl)
		} else if strings.HasPrefix(nUrl, "/") {
			e.Request.Visit(host + nUrl)
		} else if !strings.HasPrefix(nUrl, "/") {
			e.Request.Visit(url + "/" + nUrl)
		}

	})

	c.OnRequest(func(r *colly.Request) {
		currentRequests <- 1
		log.Println("Visiting", r.URL.String())
	})

	c.Visit(url)
}
