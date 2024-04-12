package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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
		log.Fatal("Error: Token not provided.")
	}

	http.HandleFunc("/task", handleTaskRequest)
	serverAddr := fmt.Sprintf("%s:%s", dashboardHost, dashboardPort)
	log.Printf("Server listening on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

func handleTaskRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 解析请求
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}
	// 处理任务
	go func() {
		response := processTask(request)
		// 将响应编码为JSON格式并发送
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}()
}

func processTask(request Request) Response {
	// 校验 Token
	if request.Token != spiderToken {
		log.Println("Invalid token received. Ignoring the task.")
		return Response{
			Token:     request.Token,
			Success:   false,
			StartTime: time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	// 记录开始时间
	startTime := time.Now()

	// 获取页面内容
	webData, success := fetchWebData(request.URL)

	// 计算运行时长
	runtime := int(time.Since(startTime).Seconds())

	// 构建响应
	loc, _ := time.LoadLocation("Asia/Shanghai")
	beijingTime := time.Now().In(loc)
	formattedTime := beijingTime.Format("2006-01-02 15:04:05")
	response := Response{
		Token:     request.Token,
		Runtime:   runtime,
		Success:   success,
		StartTime: formattedTime, // 北京时间
		WebData:   webData,
	}

	return response
}

func fetchWebData(url string) (string, bool) {
	startTime := time.Now() // 记录开始时间
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching web data: %v", err)
		return "", false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", false
	}

	log.Printf("URL: %s", resp.Request.URL)
	elapsedTime := time.Since(startTime) // 计算经过的时间
	log.Printf("Time taken: %s", elapsedTime)

	return string(body), true
}
