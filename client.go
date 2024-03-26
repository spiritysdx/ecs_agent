package main

import (
	"encoding/json"
	"fmt"
	"net"
)

type Request struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

type Response struct {
	Token   string `json:"token"`
	WebData string `json:"webdata"`
	Runtime int    `json:"runtime"`
	Success bool   `json:"success"`
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
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}

		// 解析任务
		var request Request
		err = json.Unmarshal(buffer, &request)
		if err != nil {
			fmt.Println("Error decoding request:", err.Error())
			return
		}

		// 处理任务，这里省略
		// 模拟处理任务
		response := Response{
			Token:   request.Token,
			WebData: "some_data",
			Runtime: 5,
			Success: true,
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
}
