package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"io/ioutil"
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
		fmt.Println("Error: Token not provided.")
		fmt.Println("Usage: go run your_program.go -token your_token")
		return
	}

	http.HandleFunc("/task", handleTaskRequest)
	http.ListenAndServe(dashboardHost+":"+dashboardPort, nil)
}

func handleTaskRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var request Request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		fmt.Println("Error decoding request:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 处理任务
	go func() {
		response := processTask(request)
		// 将响应编码为JSON格式并发送
		responseJSON, err := json.Marshal(response)
		if err != nil {
			fmt.Println("Error encoding response:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	}()
}

func processTask(request Request) Response {
	// 校验 Token
	if request.Token != spiderToken {
		fmt.Println("Invalid token received. Ignoring the task.")
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
		fmt.Println("Error fetching web data:", err)
		return "", false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", false
	}

	fmt.Println("URL:", resp.Request.URL)
	elapsedTime := time.Since(startTime) // 计算经过的时间
	fmt.Println("Time taken:", elapsedTime)

	return string(body), true
}
