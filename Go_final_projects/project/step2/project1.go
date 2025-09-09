package main

import (
	"bufio"
	//"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func downloadMolecule(zincID, zincVersion, zincFileType, outputDir string, wg *sync.WaitGroup) {
	defer wg.Done()
	url := fmt.Sprintf("https://zinc%s.docking.org/substances/%s.%s", zincVersion, zincID, zincFileType)
	filePath := filepath.Join(outputDir, fmt.Sprintf("%s.%s", zincID, zincFileType))

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("Skipping %s.%s (already downloaded)\n", zincID, zincFileType)
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading %s.%s: %v\n", zincID, zincFileType, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", filePath, err)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Printf("Error writing to file %s: %v\n", filePath, err)
		}
	} else {
		fmt.Printf("Failed to download %s.%s\n", zincID, zincFileType)
	}
}

func listOpener(inputIDList string) ([]string, error) {
	file, err := os.Open(inputIDList)
	if err != nil {
		return nil, fmt.Errorf("no such file %s exists", inputIDList)
	}
	defer file.Close()

	var validZincIDs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ZINC") {
			validZincIDs = append(validZincIDs, line)
		}
	}

	if len(validZincIDs) == 0 {
		return nil, fmt.Errorf("no valid ZINC IDs found in the list")
	}
	return validZincIDs, nil
}

func downloadLigands(inputIDList, outputDir string) {
	var zincVersion string
	fmt.Println("Zinc version: (choose between 15 & 20)")
	fmt.Scanln(&zincVersion)

	zincFileType := "sdf"

	zincIDList, err := listOpener(inputIDList)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Your chosen list contains %d molecules.\n", len(zincIDList))

	var wg sync.WaitGroup
	for _, zincID := range zincIDList {
		wg.Add(1)
		go downloadMolecule(zincID, zincVersion, zincFileType, outputDir, &wg)
	}
	wg.Wait()

	fmt.Println("Download job finished.")
}

/*
	func mergeSDFFiles(inputFolder, outputFileName string) {
		files, err := os.ReadDir(inputFolder)
		if err != nil {
			fmt.Printf("The folder %s does not exist.\n", inputFolder)
			return
		}

		var sdfFiles []string
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".sdf") {
				sdfFiles = append(sdfFiles, file.Name())
			}
		}

		if len(sdfFiles) == 0 {
			fmt.Println("No SDF files found in the specified folder.")
			return
		}

		outputFile, err := os.Create(outputFileName)
		if err != nil {
			fmt.Printf("Error creating output file %s: %v\n", outputFileName, err)
			return
		}
		defer outputFile.Close()

		for _, sdfFile := range sdfFiles {
			filePath := filepath.Join(inputFolder, sdfFile)
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", sdfFile, err)
				continue
			}

			lines := bytes.Split(content, []byte("\n"))
			if string(lines[len(lines)-1]) == "$$$$" {
				lines = lines[:len(lines)-1]
			}
			outputFile.Write(bytes.Join(lines, []byte("\n")))
			outputFile.Write([]byte("$$$$\n"))
		}

		fmt.Printf("All SDF files have been merged into %s\n", outputFileName)
	}
*/
func main() {
	downloadLigands("zinc_ids.txt", "set_1")
	//mergeSDFFiles("set_1", "../../3_Ligand_Preprocess/set_1/set_1.sdf")
}
