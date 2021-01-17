package main

import (
	"fmt"
	"github.com/malfunkt/iprange"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ThreadNum = 5000
	Result    *sync.Map
)

func init() {
	Result = &sync.Map{}
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func Connect(ip string, port int) (string, int, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", ip, port), time.Second*2)
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	return ip, port, err
}

func GetIpList(ips string) ([]net.IP, error) {
	address, err := iprange.ParseList(ips)
	if err != nil {
		return nil, err
	}
	list := address.Expand()
	return list, err
}

func GetPorts(selection string) ([]int, error) {
	ports := []int{}
	if selection == "" {
		return ports, nil
	}

	ranges := strings.Split(selection, ",")
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("%c[1;31m无效端口段: '%s'%c[0m", 0x1B, r, 0x1B)
			}
			part1, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("%c[1;31m无效端口号: '%s'%c[0m", 0x1B, parts[0], 0x1B)
			}
			part2, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("%c[1;31m无效端口号: '%s'%c[0m", 0x1B, parts[1], 0x1B)
			}
			if part1 > part2 {
				return nil, fmt.Errorf("%c[1;31m无效端口范围: %d-%d%c[0m", 0x1B, part1, part2, 0x1B)
			}
			for i := part1; i <= part2; i++ {
				ports = append(ports, i)
			}
		} else {
			if port, err := strconv.Atoi(r); err != nil {
				return nil, fmt.Errorf("%c[1;31m无效端口号: '%s'%c[0m", 0x1B, r, 0x1B)
			} else {
				ports = append(ports, port)
			}
		}
	}
	return ports, nil
}

func RunTask(tasks []map[string]int) {
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	for _, task := range tasks {
		for ip, port := range task {
			go func(string, int) {
				err := SaveResult(Connect(ip, port))
				_ = err
				wg.Done()
			}(ip, port)
		}
	}
	wg.Wait()
}

func GenerateTask(ipList []net.IP, ports []int) ([]map[string]int, int) {
	tasks := make([]map[string]int, 0)
	for _, ip := range ipList {
		for _, port := range ports {
			ipPort := map[string]int{ip.String(): port}
			tasks = append(tasks, ipPort)
		}
	}
	return tasks, len(tasks)
}

func AssigningTasks(tasks []map[string]int) {
	scanBatch := len(tasks) / ThreadNum
	for i := 0; i < scanBatch; i++ {
		curTask := tasks[ThreadNum*i : ThreadNum*(i+1)]
		RunTask(curTask)
	}
	if len(tasks)%ThreadNum > 0 {
		lastTasks := tasks[ThreadNum*scanBatch:]
		RunTask(lastTasks)
	}
}

func SaveResult(ip string, port int, err error) error {
	if err != nil {
		return err
	}

	v, ok := Result.Load(ip)
	if ok {
		ports, ok1 := v.([]int)
		if ok1 {
			ports = append(ports, port)
			Result.Store(ip, ports)
		}
	} else {
		ports := make([]int, 0)
		ports = append(ports, port)
		Result.Store(ip, ports)
	}
	return err
}

func PrintResult() {
	fmt.Printf("%7v               %-5v\n", "IP", "PORT")
	Result.Range(func(key, value interface{}) bool {
		fmt.Printf("%c[1;32m%-15v    %v%c[0m\n", 0x1B, key, value, 0x1B)
		return true
	})
}

func main() {
	args := os.Args
	if args == nil || len(args) < 3 {
		fmt.Printf("Useage: ./TCPPortScan 192.168.1.0/24 21,22,80-8080\n")
		return
	}

	ips, err := GetIpList(args[1])
	if err != nil {
		log.Fatal(err)
	}
	ports, err := GetPorts(args[2])
	if err != nil {
		log.Fatal(err)
	}

	task, _ := GenerateTask(ips, ports)
	AssigningTasks(task)
	PrintResult()
}
