package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	var wg sync.WaitGroup
	for i := 1; i < 65535; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			var buffer bytes.Buffer
			buffer.WriteString("127.0.0.1:")
			buffer.WriteString(strconv.Itoa(j))
			ipaddress := buffer.String()
			conn, err := net.Dial("tcp", ipaddress)
			if err != nil {
				return
			}
			conn.Close()
			fmt.Printf("%d ok \n", j)
		}(i)
	}
	wg.Wait()
	difftime := time.Since(start)
	fmt.Println("运行时间:", difftime)
}
