package main

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

func unixtime() int64 {
	return time.Now().Unix()
}

func echo(s any) {
	switch reflect.TypeOf(s).String() {
	case "string":
		fmt.Printf("%s\n", s)
	case "int", "uint", "uint32", "int32", "uint64", "int64":
		fmt.Printf("%d\n", s)
	case "[]uint8":
		fmt.Printf("%02X\n", s)
	}
}

func toStr(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case string:
		return v
	case []byte:
		//return hex.EncodeToString(v)
		return string(v)
	default:
		// Обработка других типов или возврат ошибки
		return fmt.Sprintf("%v", value)
	}
}

func IsIPv4(address string) bool {
	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}
	return ip.To4() != nil
}

// проверка на валидность ipv4 адреса..
func isValidIPv4(ip string) bool {
	re := regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return re.MatchString(ip)
}

// разбираем на интерфейс, адрес и порт (1234)
func parseUdpAddr(udpAddr string) (*net.Interface, string, int, error) {
	re := regexp.MustCompile(`^udp://([^@]*)@([0-9.]+)(?::(\d+))?$`)
	matches := re.FindStringSubmatch(udpAddr)
	if len(matches) != 4 || !isValidIPv4(matches[2]) {
		return nil, "", 0, errors.New("Invalid address format: " + udpAddr)
	}
	ifi, err := net.InterfaceByName(matches[1])
	if matches[1] == "" || err != nil {
		ifi = nil
	}
	port, err := strconv.Atoi(matches[3])
	if err != nil || (port < 100 || port > 65535) {
		port = 1234
	}
	return ifi, matches[2], port, nil
}
