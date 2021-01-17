package main

import (
	"fmt"
	"github.com/malfunkt/iprange"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func Connect(ip string, port int) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", ip, port), time.Second*2)
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	return conn, err
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

func main() {
	args := os.Args
	if args == nil || len(args) < 3 {
		fmt.Printf("Useage: ./TCPPortScan 192.168.1.1/24 21,22,80-8080\n")
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
	fmt.Printf("%7v           %-5v\n", "IP", "PORT")
	for _, ip := range ips {
		for _, port := range ports {
			_, err = Connect(ip.String(), port)
			if err != nil {
				continue
			}
			fmt.Printf("%c[1;32m%-15v    %-5v%c[0m\n", 0x1B, ip, port, 0x1B)
		}
	}
}
