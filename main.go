package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//处理ip池
var ipaddress chan string = make(chan string)

//线程数
var thread chan int = make(chan int)
var threadend chan int = make(chan int)

type inputParams struct {
	ip     string
	port   string
	thread int
}

func handelParams(p inputParams) {
	ips := handelIP(p.ip)
	if len(ips) == 0 {
		fmt.Println("error:ip参数格式错误")
		flag.Usage()
		os.Exit(0)
	}

	ports := handelPorts(p.port)
	if len(ports) == 0 {
		fmt.Println("error:port参数格式错误")
		flag.Usage()
		os.Exit(0)
	}
	//传送启动线程数
	//最大线程2048
	t := p.thread

	if p.thread < 1 {
		t = 1
	} else if p.thread > 2048 {
		t = 2048
	}
	thread <- t
	//fmt.Println(len(ips))
	for _, ip := range ips {
		for _, port := range ports {
			ipaddress <- ip + ":" + strconv.Itoa(port)
		}
	}

}

func handelPorts(s string) []int {
	var ports = make([]int, 0)
	if strings.Index(s, ",") != -1 {
		tmpPorts := strings.Split(s, ",")
		for _, v := range tmpPorts {
			port, err := strconv.Atoi(v)
			if err != nil {
				//端口不合法
				//fmt.Println("'port' Setting error")
				return make([]int, 0)
			}
			ports = append(ports, port)
		}
		return ports
	} else if strings.Index(s, "~") != -1 {
		tmpPorts := strings.Split(s, "~")
		var startPort, endPort int
		var err error
		startPort, err = strconv.Atoi(tmpPorts[0])
		if err != nil || startPort < 1 || startPort > 65535 {
			//开始端口不合法
			return ports
		}
		if len(tmpPorts) >= 2 {
			//指定结束端口
			endPort, err = strconv.Atoi(tmpPorts[1])
			if err != nil || endPort < 1 || endPort > 65535 || endPort < startPort {
				//结束端口不合法
				return ports
			}
		} else {
			//未指定结束端口
			endPort = 65535
		}
		for i := 0; startPort+i <= endPort; i++ {
			ports = append(ports, startPort+i)
		}
	}
	return ports
}

func handelIP(s string) []string {
	var ips = make([]string, 0)
	address := net.ParseIP(s)
	// 单个IP处理
	if address != nil {
		ips = append(ips, s)
		return ips
	}
	//指定IP处理
	if strings.Index(s, ",") != -1 {
		tmpIPs := strings.Split(s, ",")
		for _, v := range tmpIPs {
			tmpAddress := net.ParseIP(v)
			if tmpAddress != nil {
				ips = append(ips, v)
				continue
			}
			return make([]string, 0)
		}
	}
	//ip段处理
	if strings.Index(s, "~") != -1 {
		tmpIPs := strings.Split(s, "~")
		if len(tmpIPs) != 2 {
			return make([]string, 0)
		}
		for _, v := range tmpIPs {

			tmpAddress := net.ParseIP(v)
			if tmpAddress == nil {
				//fmt.Println(v)
				return make([]string, 0)
			}
		}
		startIP := tmpIPs[0]
		endIP := tmpIPs[1]
		for ; startIP != endIP; startIP = nextIP(startIP) {
			if startIP != "" {
				ips = append(ips, startIP)
			}
		}
		ips = append(ips, endIP)
	}
	return ips
}

func nextIP(ip string) string {
	ips := strings.Split(ip, ".")
	var i int
	for i = len(ips) - 1; i >= 0; i-- {
		n, _ := strconv.Atoi(ips[i])
		if n >= 255 {
			//进位
			if i == 3 {
				ips[i] = "1"
			} else {
				ips[i] = "0"
			}
		} else {
			//+1
			n++
			ips[i] = strconv.Itoa(n)
			break
		}
	}
	if i == -1 {
		//全部IP段都进行了进位,说明此IP本身已超出范围
		return ""
	}
	var nIP bytes.Buffer
	leng := len(ips)
	for j := 0; j < leng; j++ {
		if j == leng-1 {
			nIP.WriteString(ips[j])
		} else {
			nIP.WriteString(ips[j])
			nIP.WriteString(".")
		}
	}
	return nIP.String()
}

func scan(i int) {
	address, ok := <-ipaddress
	for ok {
		conn, err := net.Dial("tcp", address)
		if err == nil {
			//端口开放
			fmt.Printf("%s is ok--%d\n", address, i)
			conn.Close()
		} else {
			//fmt.Printf("%s is err-%d\n", address, i)
		}

		//fmt.Printf("%s is error--%d\n", address, i)
		address, ok = <-ipaddress
	}
	thread <- 0
}

func runScan() {

	t, ok := <-thread
	nowThread := t
	if ok {
		for i := 0; i < nowThread; i++ {
			go scan(i)
		}
	}
	//等待线程终止
	for <-thread == 0 {
		nowThread--
		fmt.Println("-")
		if nowThread == 0 {
			//全部线程已终止,关闭结果写入,退出程序
			//close(result)
			break
		}
	}
}

func main() {
	startTime := time.Now()
	var params inputParams
	flag.StringVar(&params.ip, "ip", "", "ip 192.168.1.1	192.168.1.1,192.168.1.2		192.168.1.1~192.168.1.80")
	flag.StringVar(&params.port, "port", "20,21,22,80,443,3306", "80	80,81,82	80~90")
	flag.IntVar(&params.thread, "thread", 1024, "启动线程数")
	flag.Parse()

	go runScan()
	handelParams(params) //参数处理
	needtime := time.Since(startTime)
	fmt.Printf("run time %s", needtime)
}
