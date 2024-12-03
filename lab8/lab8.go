/*
package main

import (

	"flag"
	"github.com/gocolly/colly"

)

func main() {

		flag.Parse()

		c := colly.NewCollector()

		c.OnHTML("???", func(e *colly.HTMLElement) {
		})
	}
*/
package main

import (
	"flag"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"strings"
)

func main() {
	// 定義 max flag，預設最多印出 10 條留言
	max := flag.Int("max", 10, "Max number of comments to show")
	flag.Parse()

	// 建立 Colly 收集器
	c := colly.NewCollector()

	// 計數器，記錄已處理的留言數量
	counter := 0

	// 定義爬取 .push 節點的處理方式
	c.OnHTML(".push", func(e *colly.HTMLElement) {
		if counter >= *max {
			return
		}

		// 提取留言者名稱、留言內容和時間
		username := e.ChildText(".push-userid")
		comment := e.ChildText(".push-content")
		rawTime := e.ChildText(".push-ipdatetime")

		// 處理 comment（去掉前面的 ": "）
		if len(comment) > 2 {
			comment = strings.TrimSpace(comment[2:])
		}

		// 處理時間（只提取時間部分）
		timestamp := strings.TrimSpace(rawTime)

		// 如果 comment 非空，則印出資訊
		if comment != "" {
			counter++
			fmt.Printf("%d. 名字：%s，留言: %s，時間： %s\n", counter, username, comment, timestamp)
		}
	})

	// 定義錯誤處理回調
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error:", err)
	})

	// 訪問 PTT 頁面
	err := c.Visit("https://www.ptt.cc/bbs/joke/M.1481217639.A.4DF.html")
	if err != nil {
		log.Fatal("Failed to visit the page:", err)
	}
}
