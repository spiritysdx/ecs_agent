package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/parnurzeal/gorequest"
)

type Request struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

type Response struct {
	Token     string `json:"token"`
	Runtime   int    `json:"runtime"`
	StartTime string `json:"start_time"`
	Success   bool   `json:"success"`
	WebData   string `json:"webdata,omitempty"`
}

var spiderToken, dashboardHost, dashboardPort string

func main() {
	flag.StringVar(&spiderToken, "token", "", "爬虫校验的Token")
	flag.StringVar(&dashboardHost, "host", "", "主控的IP地址")
	flag.StringVar(&dashboardPort, "port", "", "主控的通信端口")
	flag.Parse()
	if spiderToken == "" {
		fmt.Println("Error: Token not provided.")
		fmt.Println("Usage: go run your_program.go -token your_token")
		os.Exit(1)
	}
	for {
		conn, err := net.Dial("tcp", dashboardHost+":"+dashboardPort)
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			time.Sleep(6 * time.Second)
			continue
		}
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		// 读取消息长度
		lenBuf := make([]byte, 4)
		_, err := conn.Read(lenBuf)
		if err != nil {
			fmt.Println("Error reading message length:", err.Error())
			return
		}
		msgLen := binary.BigEndian.Uint32(lenBuf)
		// 读取消息内容
		buffer := make([]byte, msgLen)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading message:", err.Error())
			return
		}
		var request Request
		err = json.Unmarshal(buffer[:n], &request)
		if err != nil {
			fmt.Println("Error decoding request:", err.Error())
			continue
		}
		go handleTask(conn, request)
	}
}

func handleTask(conn net.Conn, request Request) {
	if request.Token != spiderToken {
		fmt.Println("Invalid token received. Ignoring the task.")
		return
	}
	startTime := time.Now()
	webData, success := fetchWebData(request.URL)
	runtime := int(time.Since(startTime).Seconds())
	loc, _ := time.LoadLocation("Asia/Shanghai")
	beijingTime := time.Now().In(loc)
	formattedTime := beijingTime.Format("2006-01-02 15:04:05")
	response := Response{
		Token:     request.Token,
		Runtime:   runtime,
		Success:   success,
		StartTime: formattedTime,
		WebData:   webData,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error encoding response:", err.Error())
		return
	}
	// 发送消息长度
	msgLen := uint32(len(responseJSON))
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, msgLen)
	_, err = conn.Write(lenBuf)
	if err != nil {
		fmt.Println("Error sending message length:", err.Error())
		return
	}
	// 发送消息内容
	_, err = conn.Write(responseJSON)
	if err != nil {
		fmt.Println("Error sending response:", err.Error())
		return
	}
	fmt.Println("Sent response size:", len(responseJSON))
}

func fetchWebData(url string) (string, bool) {
	startTime := time.Now()
	request := gorequest.New()
	resp, body, err := request.Get(url).End()
	if err != nil {
		fmt.Printf("Error reading response body: %v \n", url)
		return "", false
	}
	fmt.Println("URL:", resp.Request.URL)
	elapsedTime := time.Since(startTime)
	fmt.Println("Time taken:", elapsedTime)
	return body, true
}
