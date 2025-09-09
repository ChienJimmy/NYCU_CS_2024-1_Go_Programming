package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 設定Zinc ID檔案目錄
const zincIDsDir = `zinc_ids`
const resultFileName = "zinc_ids.txt" // 結果檔案名稱

func main() {
	// 刪除舊的 zinc_ids.txt 檔案（如果存在）
	if err := os.Remove(resultFileName); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to remove old result file: %v", err)
	}

	// 設定路由
	http.HandleFunc("/", serveForm)
	http.HandleFunc("/process", processRequest)

	// 啟動伺服器
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// 伺服器主頁，提供HTML表單
func serveForm(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

// 處理用戶提交的表單
func processRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 解析表單數據
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// 獲取條件
	logPValues := r.Form["logP[]"]
	molecularWeights := r.Form["molecularWeight[]"]
	quantities := r.Form["quantity[]"]

	if len(logPValues) == 0 || len(logPValues) > 5 {
		http.Error(w, "Conditions must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// 準備輸出結果檔案
	output, err := os.Create(resultFileName)
	if err != nil {
		http.Error(w, "Failed to create result file", http.StatusInternalServerError)
		return
	}
	defer output.Close()

	// 處理每組條件
	for i := range logPValues {
		logP := logPValues[i]
		molecularWeight := molecularWeights[i]
		quantity, _ := strconv.Atoi(quantities[i])

		// 構造文件名稱
		fileName := fmt.Sprintf("zinc_ids_%s%s.txt", molecularWeight, logP)
		filePath := filepath.Join(zincIDsDir, fileName)

		// 打開對應檔案
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to open file: %s", fileName), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// 讀取Zinc ID
		var zincIDs []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			zincIDs = append(zincIDs, scanner.Text())
		}

		if len(zincIDs) < quantity {
			http.Error(w, fmt.Sprintf("Not enough Zinc IDs in file: %s", fileName), http.StatusInternalServerError)
			return
		}

		// 隨機選取Zinc ID
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(zincIDs), func(i, j int) { zincIDs[i], zincIDs[j] = zincIDs[j], zincIDs[i] })

		selected := zincIDs[:quantity]
		output.WriteString(strings.Join(selected, "\n") + "\n")
	}

	// 使用模板渲染結果頁面
	tmpl := template.Must(template.ParseFiles("completion.html"))
	data := struct {
		FilePath string
	}{
		FilePath: resultFileName, // 假設輸出的文件名是 zinc_ids.txt
	}
	tmpl.Execute(w, data)
}
