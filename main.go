package main

import (
	"fmt"
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
		fmt.Printf("Current Connections: %d - Current Requests: %d - The secons: %d\n", total, totalRequests, seconds)
		select {
		case <-currentConnections:
			total++
		case <-disconnections:
			total--
		case <-currentRequests:
			totalRequests++
		}
	}
}

func increaseSeconds() {
	for {
		time.Sleep(1 * time.Second)
		seconds++
		if seconds == ttl {
			os.Exit(0)
		}
	}
}

func main() {
	currentConnections = make(chan int, 10)
	disconnections = make(chan int, 10)
	currentRequests = make(chan int, 10)
	go display()
	go increaseSeconds()
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

	for i := 0; i < numberOfTask; i++ {
		wg.Add(1)
		go func(link string) {
			defer func() {
				disconnections <- 1
				wg.Done()
			}()
			currentConnections <- 1
			collect(link)
		}(url)
		time.Sleep(time.Second)
	}

	ttl = seconds + 300

	wg.Wait()
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
		} else if strings.HasPrefix(nUrl, "/") == false {
			e.Request.Visit(url + "/" + nUrl)
		}

	})
	c.OnRequest(func(r *colly.Request) {
		currentRequests <- 1
	})

	c.Visit(url)
}
