package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// Struct to store user input
type InputData struct {
	Conditions []Condition
	Message    string
}

type Condition struct {
	LogP            string
	MolecularWeight string
	ZincIDs         []string
	FileName        string
}

var tpl = template.Must(template.ParseFiles("index.html"))

func main() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/fetch", fetchZincIDs)
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	data := InputData{}
	tpl.Execute(w, data)
}

// fetchZincIDs 处理抓取 ZINC IDs
func fetchZincIDs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 删除旧的txt文件
	/*err := deleteOldTxtFiles()
	if err != nil {
		log.Printf("删除旧文件时发生错误: %v", err)
	}

	// 解析表单数据
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "无法解析表单", http.StatusInternalServerError)
		return
	}*/

	// 获取用户输入的条件
	var conditions []Condition
	for i := 0; i < 5; i++ {
		logP := r.FormValue(fmt.Sprintf("logp%d", i+1))
		molecularWeight := r.FormValue(fmt.Sprintf("molecularweight%d", i+1))
		if logP != "" && molecularWeight != "" {
			conditions = append(conditions, Condition{
				LogP:            logP,
				MolecularWeight: molecularWeight,
			})
		}
	}

	if len(conditions) == 0 {
		data := InputData{Message: "至少需要输入一组条件。"}
		tpl.Execute(w, data)
		return
	}

	// 抓取ZINC ID
	var wg sync.WaitGroup
	for i, cond := range conditions {
		wg.Add(1)
		go func(i int, cond Condition) {
			defer wg.Done()

			// 映射输入为ZINC20网址字母
			logPMap := map[string]string{"-1": "A", "0": "B", "1": "C", "2": "D", "2.5": "E", "3": "F", "3.5": "G", "4": "H", "4.5": "I", "5": "J", ">5": "K"}
			mwMap := map[string]string{"200": "A", "250": "B", "300": "C", "325": "D", "350": "E", "375": "F", "400": "G", "425": "H", "450": "I", "500": "J", ">500": "K"}

			logPLetter := logPMap[cond.LogP]
			mwLetter := mwMap[cond.MolecularWeight]

			if logPLetter == "" || mwLetter == "" {
				return
			}

			// 生成正确的基础URL
			baseURL := fmt.Sprintf("https://zinc20.docking.org/substances/subsets/%s%s/", mwLetter, logPLetter)
			fmt.Println("URL = ", baseURL)

			// 抓取ZINC ID
			var zincIDs []string
			var pageWg sync.WaitGroup
			// 使用 goroutine 同时抓取多个页面
			for page := 1; page <= 50; page++ {
				pageWg.Add(1)
				go func(page int) {
					defer pageWg.Done()
					pageURL := fmt.Sprintf("%s?page=%d", baseURL, page)
					fmt.Println("URL = ", pageURL)
					ids, err := scrapeZincIDs(pageURL)
					if err != nil {
						log.Printf("Error scraping page %d for %s_%s: %v", page, cond.MolecularWeight, cond.LogP, err)
						return
					}
					zincIDs = append(zincIDs, ids...)
				}(page)
			}
			pageWg.Wait()

			// 格式化ZINC IDs（去除前导零并确保12位）
			formattedZincIDs := make([]string, 0)
			for _, id := range zincIDs {
				formattedID := formatZincID(id)
				if formattedID != "" {
					formattedZincIDs = append(formattedZincIDs, formattedID)
				}
			}

			// 对ZINC ID按照数字排序
			sort.Slice(formattedZincIDs, func(i, j int) bool {
				numI := extractNumericPart(formattedZincIDs[i])
				numJ := extractNumericPart(formattedZincIDs[j])
				return numI < numJ
			})

			// 保存ZINC ID到文件
			err := saveToFileWithLetters(logPLetter, mwLetter, formattedZincIDs)
			if err != nil {
				log.Printf("Error saving to file %s%s: %v", mwLetter, logPLetter, err)
				return
			}

			// 更新条件的ZincIDs
			cond.ZincIDs = formattedZincIDs
			conditions[i] = cond
		}(i, cond)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 显示抓取结果
	data := InputData{
		Conditions: conditions,
		Message:    "抓取完成，ZINC ID 已保存至对应文件。",
	}
	tpl.Execute(w, data)
}

// 删除旧的txt文件
/*func deleteOldTxtFiles() error {
	// 获取当前目录中的所有文件
	files, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("无法读取目录: %v", err)
	}

	// 遍历文件列表，删除以 ".txt" 结尾的文件
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			err := os.Remove(file.Name())
			if err != nil {
				return fmt.Errorf("删除文件 %s 失败: %v", file.Name(), err)
			}
		}
	}

	return nil
}*/

// 将抓取到的ZINC IDs保存到文件，文件名使用字母表示 LogP 和 Molecular Weight
func saveToFileWithLetters(logPLetter, mwLetter string, zincIDs []string) error {
	fileName := fmt.Sprintf("zinc_ids_%s%s.txt", mwLetter, logPLetter)
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入ZINC IDs
	for _, id := range zincIDs {
		_, err := file.WriteString(id + "\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}

	return nil
}

// scrapeZincIDs 从网页中抓取ZINC ID
func scrapeZincIDs(url string) ([]string, error) {
	// 请求页面内容
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取页面失败: %v", err)
	}
	defer res.Body.Close()

	// 使用 goquery 解析HTML
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	// 抓取所有ZINC ID
	var zincIDs []string
	doc.Find(".zinc-id.caption").Each(func(i int, s *goquery.Selection) {
		// 获取ZINC ID的文本
		id := s.Text()
		id = strings.TrimSpace(id)

		// 确保只包含ZINC ID，不包括其他内容
		if strings.HasPrefix(id, "ZINC") {
			// 去除后面的描述信息，只保留ZINC ID部分
			id = strings.Split(id, " ")[0]
			zincIDs = append(zincIDs, id)
		}
	})

	return zincIDs, nil
}

// 格式化ZINC ID：去除前导零并确保12位
func formatZincID(id string) string {
	// Trim whitespace
	id = strings.TrimSpace(id)

	// Ensure the ID starts with "ZINC"
	if strings.HasPrefix(id, "ZINC") {
		// Get the number part after "ZINC"
		numPart := id[4:]

		// 如果是 7 位数字，补充为 12 位
		if len(numPart) < 12 {
			for len(numPart) < 12 {
				numPart = "0" + numPart
			}
		}

		// Return the properly formatted ID
		return "ZINC" + numPart
	}
	return ""
}

// 提取ZINC ID中的数字部分
func extractNumericPart(id string) int {
	// 从 ZINC ID 提取数字部分并转为整数
	var numericPart string
	for _, char := range id[4:] {
		if char >= '0' && char <= '9' {
			numericPart += string(char)
		}
	}
	num, err := strconv.Atoi(numericPart)
	if err != nil {
		return 0
	}
	return num
}
