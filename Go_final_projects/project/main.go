package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// runProjectGo 用來執行 project.go，並將其作為後台進程運行
func runProjectGo() {
	cmd := exec.Command("go", "run", "step1/project.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 開始執行並讓它在背景運行
	err := cmd.Start()
	if err != nil {
		log.Fatalf("執行 project.go 失敗: %v", err)
	}

	// 顯示訊息並提示 project.go 正在運行
	fmt.Println("project.go 正在運行，等待 zinc_ids.txt 被生成...")
}

// runProject1Go 用來執行 project1.go 將 ID 轉換為 sdf 檔案
func runProject1Go() {
	// 每次運行 project1.go 前，清空 set_1 資料夾
	clearOutputDir("set_1")

	cmd := exec.Command("go", "run", "step2/project1.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("執行 project1.go 失敗: %v", err)
	}
	fmt.Println("project1.go 運行完成！")
}

// clearOutputDir 清空 set_1 資料夾
func clearOutputDir(outputDir string) {
	files, err := os.ReadDir(outputDir)
	if err != nil {
		fmt.Printf("無法讀取資料夾 %s: %v\n", outputDir, err)
		return
	}

	for _, file := range files {
		filePath := filepath.Join(outputDir, file.Name())
		err := os.RemoveAll(filePath)
		if err != nil {
			fmt.Printf("無法刪除檔案 %s: %v\n", filePath, err)
		} else {
			fmt.Printf("已刪除 %s\n", filePath)
		}
	}
}

// watchZincIdsFile 用來監控 zinc_ids.txt 檔案的變動
func watchZincIdsFile(fileName string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("無法建立文件監視器: %v", err)
	}
	defer watcher.Close()

	// 開始監控 zinc_ids.txt 所在的目錄
	dirPath := "./"
	err = watcher.Add(dirPath)
	if err != nil {
		log.Fatalf("無法監控目錄: %v", err)
	}

	fmt.Printf("正在監控 %s 目錄中的 %s 檔案...\n", dirPath, fileName)

	// 避免重複執行，可以設置檔案變動的冷卻時間
	var lastWriteTime time.Time

	// 無窮迴圈監視文件變動
	for {
		select {
		case event := <-watcher.Events:
			// 顯示事件操作
			fmt.Printf("檢測到事件: %v\n", event)

			// 只處理與檔案相關的事件
			if event.Name == fileName { // 改為只比對檔案名稱
				// 檔案被刪除
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Printf("檔案已刪除: %s\n", event.Name)
					// 等待檔案被重新創建
					fmt.Println("等待檔案重建...")
					continue // 這會讓程式進入等待狀態，繼續監控檔案創建
				}

				// 檔案被創建
				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Printf("檔案已創建: %s\n", event.Name)

					// 等待檔案創建完成後，稍微延遲，確保檔案已經有內容
					time.Sleep(2 * time.Second)

					// 確保檔案已經有內容，再執行 project1.go
					if _, err := os.Stat(fileName); err == nil {
						// 檢查檔案大小來確認是否有內容
						fileInfo, err := os.Stat(fileName)
						if err == nil && fileInfo.Size() > 0 {
							fmt.Println("檔案有內容，開始執行 project1.go")
						} else {
							fmt.Println("檔案沒有內容，等待檔案填充...")
						}
					}
				}

				// 檔案被修改
				if event.Op&fsnotify.Write == fsnotify.Write {
					// 如果文件在冷卻期內，跳過此次事件
					if time.Since(lastWriteTime) < 2*time.Second {
						// 不處理過於頻繁的寫入事件
						fmt.Println("檔案寫入過於頻繁，等待下次處理...")
						continue
					}
					// 更新最後的寫入時間
					lastWriteTime = time.Now()

					// 可以選擇在檔案變動時處理
					fmt.Printf("檔案已修改: %s\n", event.Name)
					runProject1Go()
				}
			}
		case err := <-watcher.Errors:
			log.Printf("監控錯誤: %v", err)
		}
	}
}

func main() {
	// 先運行 project.go 抓取資料，並讓它在背景執行
	runProjectGo()

	// 監控 zinc_ids.txt 檔案
	fileName := "zinc_ids.txt"
	// 開始監控檔案變動
	watchZincIdsFile(fileName)
}
