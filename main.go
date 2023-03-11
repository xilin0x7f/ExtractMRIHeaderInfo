/*
@Author: xilin0x7f, https://github.com/xilin0x7f
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	root := flag.String("root", "", "数据存放路径，root/组别/被试")
	outputFileName := flag.String("o", "HeaderInfo.xlsx", "")
	strReg := flag.String("strReg", "nii", "")
	flag.Parse()
	strReg4Match := regexp.MustCompile(*strReg)
	groups, _ := os.ReadDir(*root)
	var jsonFilesName []string
	for _, group := range groups {
		subjects, _ := os.ReadDir(filepath.Join(*root, group.Name()))
		for _, subject := range subjects {
			subjectFiles, _ := os.ReadDir(filepath.Join(*root, group.Name(), subject.Name()))
			for _, subjectFile := range subjectFiles {
				if strReg4Match.MatchString(subjectFile.Name()) {
					strSplit := strings.Split(subjectFile.Name(), ".")
					if strSplit[len(strSplit)-1] != "nii" && strSplit[len(strSplit)-1] != "gz" {
						continue
					}
					out, err := exec.Command("mrinfo", filepath.Join(*root, group.Name(), subject.Name(), subjectFile.Name()), "-json_all",
						filepath.Join(*root, group.Name(), subject.Name(), subjectFile.Name()+".json")).CombinedOutput()
					outStr := string(out)
					fmt.Println(outStr)
					if err != nil {
						fmt.Println(err)
					}

					jsonFilesName = append(jsonFilesName, filepath.Join(*root, group.Name(), subject.Name(), subjectFile.Name()+".json"))
				}
			}
		}
	}
	if err := WriteJson2XLSX(jsonFilesName, filepath.Join(*root, *outputFileName), "Sheet1", "A"); err != nil {
		log.Fatal(err)
	}
	for _, jsonFileName := range jsonFilesName {
		_ = os.Remove(jsonFileName)
	}
}

func WriteJson2XLSX(filesName []string, dstFileName, sheetName, start string) error {
	keysMap := make(map[string]int)
	for _, fileName := range filesName {
		file, _ := os.Open(fileName)
		var jsonData map[string]interface{}
		reader := io.Reader(file)
		decoder := json.NewDecoder(reader)
		_ = decoder.Decode(&jsonData)
		_ = file.Close()
		for key := range jsonData {
			keysMap[key]++
		}
	}
	var keys []string
	for key := range keysMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	resMap := make(map[string][]interface{})
	for _, fileName := range filesName {
		file, _ := os.Open(fileName)
		var jsonData map[string]interface{}
		reader := io.Reader(file)
		decoder := json.NewDecoder(reader)
		_ = decoder.Decode(&jsonData)
		_ = file.Close()
		for _, key := range keys {
			resMap[key] = append(resMap[key], jsonData[key])
		}
	}
	res := make([][]interface{}, len(resMap[keys[0]])+1)
	for idx := range res {
		res[idx] = make([]interface{}, len(keys)+1)
	}
	res[0][0] = ""
	for idx, key := range keys {
		res[0][idx+1] = key
	}
	for rowIdx := range resMap[keys[0]] {
		res[rowIdx+1][0] = filesName[rowIdx]
		for colIdx := range keys {
			res[rowIdx+1][colIdx+1] = resMap[keys[colIdx]][rowIdx]
		}
	}
	err := Write2XLSX(dstFileName, sheetName, start, res)
	return err
}
func Write2XLSX(fileName, sheetName, start string, data [][]interface{}) error {
	f := excelize.NewFile()
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)
	for idx := range data {
		if err := f.SetSheetRow(sheetName, fmt.Sprint(start, idx+1), &data[idx]); err != nil {
			return err
		}
	}
	if err := f.SaveAs(fileName); err != nil {
		return err
	}
	return nil
}
