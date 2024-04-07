package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"github.com/parnurzeal/gorequest"
	"os"
	"time"
)

// go run client.go -token 

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

var SpidersToken string

func main() {
	// 指定校验的token明文
	flag.StringVar(&SpidersToken, "token", "", "Token for authentication")
	flag.Parse()
	if SpidersToken == "" {
		fmt.Println("Error: Token not provided.")
		fmt.Println("Usage: go run your_program.go -token your_token")
		os.Exit(1)
	}
	for {
		// 连接服务端
		conn, err := net.Dial("tcp", "localhost:7788")
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			time.Sleep(6 * time.Second) // 等待 6 秒后尝试重新连接
			continue
		}
		// 连接成功后开始处理任务
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		// 接收服务端发送的任务
		buffer := make([]byte, 409600)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return // 如果出错就退出，等待外层循环重新连接
		}
		// 解析任务
		var request Request
		err = json.Unmarshal(buffer[:n], &request)
		if err != nil {
			fmt.Println("Error decoding request:", err.Error())
			continue // 继续等待下一个任务
		}
		// 处理任务
		go handleTask(conn, request)
	}
}

func handleTask(conn net.Conn, request Request) {
	// 校验 Token
	if request.Token != SpidersToken {
		fmt.Println("Invalid token received. Ignoring the task.")
		return
	}
	// 记录开始时间
	startTime := time.Now()
	// 获取页面内容
	webData, success := fetchWebData(request.URL)
	// 计算运行时长
	runtime := int(time.Since(startTime).Seconds())
	// 构建响应
	response := Response{
		Token:     request.Token,
		Runtime:   runtime,
		Success:   success,
		StartTime: startTime.Format(time.RFC3339), // 北京时间
		WebData:   webData,
	}
	// 将响应编码为JSON格式
	responseJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error encoding response:", err.Error())
		return
	}
	// 发送响应给服务端
	_, err = conn.Write(responseJSON)
	if err != nil {
		fmt.Println("Error sending response:", err.Error())
		return
	}
}

func fetchWebData(url string) (string, bool) {
	startTime := time.Now() // 记录开始时间
	request := gorequest.New()
	resp, body, err := request.Get(url).End()
	if err != nil {
		fmt.Printf("Error reading response body: %v \n", url)
		return "", false
	}
	fmt.Println("URL:", resp.Request.URL)
	elapsedTime := time.Since(startTime) // 计算经过的时间
	fmt.Println("Time taken:", elapsedTime)
	return body, true
}
