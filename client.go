package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Request struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

type Response struct {
	Token     string `json:"token"`
	WebData   string `json:"webdata,omitempty"`
	Runtime   int    `json:"runtime"`
	StartTime string `json:"start_time"`
	Success   bool   `json:"success"`
}

func main() {
	// 连接服务端
	conn, err := net.Dial("tcp", "localhost:7788")
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	defer conn.Close()

	for {
		// 接收服务端发送的任务
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			continue // 继续等待下一个任务
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
	// 记录开始时间
	startTime := time.Now()

	// 获取页面内容
	webData, success := fetchWebData(request.URL)

	// 计算运行时长
	runtime := int(time.Since(startTime).Seconds())

	// 构建响应
	response := Response{
		Token:     request.Token,
		WebData:   webData,
		Runtime:   runtime,
		StartTime: startTime.Format(time.RFC3339), // 北京时间
		Success:   success,
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
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error fetching web data:", err.Error())
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching web data. Status code: %d\n", resp.StatusCode)
		return "", false
	}

	// 读取页面内容
	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil {
		fmt.Println("Error reading response body:", err.Error())
		return "", false
	}

	return string(body[:n]), true
}
