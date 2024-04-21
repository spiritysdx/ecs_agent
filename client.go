package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"io"
	"net/http"
	"os"
	"time"
)

type CrawlerTask struct {
	Token       string `json:"token"`        // 校验密码
	Tag         string `json:"tag"`          // 商家TAG
	URL         string `json:"url"`          // 需要爬取的链接
	BillingType string `json:"billing_type"` // 爬取的类型
	CrawlNum    int    `json:"crawl_num"`    // 包含的商品个数
	ExtraHeader string `json:"extra_header"` // 额外的请求头
	ReqMethod   string `json:"req_method"`   // 请求模式
}

type TaskFromData struct {
	Data CrawlerTask `json:"data"`
}

type CrawlerResult struct {
	Token       string `json:"token"`             // 校验密码
	Tag         string `json:"tag"`               // 商家TAG
	URL         string `json:"url"`               // 需要爬取的链接
	BillingType string `json:"billing_type"`      // 爬取的类型
	CrawlNum    int    `json:"crawl_num"`         // 包含的商品个数
	Runtime     int    `json:"runtime"`           // 爬虫耗时
	StartTime   string `json:"start_time"`        // 爬虫开始时间
	Success     bool   `json:"success"`           // 是否成功抓取页面
	ReqMethod   string `json:"req_method"`        // 请求模式
	WebData     string `json:"webdata,omitempty"` // 页面的html源码
}

var spiderToken, dashboardHost, dashboardPort string

func getOneTask(endpoint string) (CrawlerTask, error) {
	url := "http://" + dashboardHost + ":" + dashboardPort + endpoint
	data := map[string]string{"token": spiderToken}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return CrawlerTask{}, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return CrawlerTask{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return CrawlerTask{}, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CrawlerTask{}, err
	}
	var task TaskFromData
	err = json.Unmarshal(body, &task)
	if err != nil {
		return CrawlerTask{}, err
	}
	return task.Data, nil
}

func getSpidersOneTask() (CrawlerTask, error) {
	return getOneTask("/spiders/getonetask")
}

func getFASpidersOneTask() (CrawlerTask, error) {
	return getOneTask("/AllPdSpiders/getonetask")
}

func fetchWebData(url string) (string, bool) {
	startTime := time.Now()
	request := gorequest.New()
	req := request.Get(url).Retry(3, 6*time.Second, http.StatusBadRequest, http.StatusInternalServerError)
	req.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36")
	req.Set("Accept",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	_, body, err := req.End()
	if err != nil {
		fmt.Printf("Error reading response body: %v \n", url)
		return "", false
	}
	fmt.Println("URL:", url)
	elapsedTime := time.Since(startTime)
	fmt.Println("Time taken:", elapsedTime)
	return body, true
}

func handleTask(task CrawlerTask, endpoint string) {
	if task.Token != spiderToken {
		fmt.Println("Invalid token received. Ignoring the task.")
		return
	}
	if task.URL == "" || task.Tag == "" {
		fmt.Println("Invalid URL or Tag. Ignoring the task.")
		return
	}
	startTime := time.Now()
	webData, success := fetchWebData(task.URL)
	runtime := int(time.Since(startTime).Seconds())
	loc, _ := time.LoadLocation("Asia/Shanghai")
	beijingTime := time.Now().In(loc)
	formattedTime := beijingTime.Format("2006-01-02 15:04:05")
	response := CrawlerResult{
		Token:       spiderToken,
		Tag:         task.Tag,
		CrawlNum:    task.CrawlNum,
		BillingType: task.BillingType,
		URL:         task.URL,
		Runtime:     runtime,
		Success:     success,
		StartTime:   formattedTime,
		ReqMethod:   task.ReqMethod,
		WebData:     webData,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error encoding response:", err.Error())
		return
	}
	// 发送信息
	url := "http://" + dashboardHost + ":" + dashboardPort + endpoint
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(responseJSON))
	if err != nil {
		fmt.Println("Error sending post:", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP request failed with status code: %d", resp.StatusCode)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error readall body: %v", err.Error())
		return
	}
	fmt.Println("Sent response result:", string(body))
}

func handleSpidersOneTask(task CrawlerTask) {
	handleTask(task, "/spiders/handletask")
}

func handleFASpidersOneTask(task CrawlerTask) {
	handleTask(task, "/AllPdSpiders/handletask")
}

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
		task1, err1 := getFASpidersOneTask()
		if err1 != nil {
			fmt.Println("Error getting task:", err1.Error())
		} else {
			go handleFASpidersOneTask(task1)
		}
		task2, err2 := getSpidersOneTask()
		if err2 != nil {
			fmt.Println("Error getting task:", err2.Error())
			time.Sleep(6 * time.Second)
			continue
		} else {
			go handleSpidersOneTask(task2)
		}
		time.Sleep(1 * time.Second)
	}
}
