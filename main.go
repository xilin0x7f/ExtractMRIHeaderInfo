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
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
)

func displayHelp() {
	fmt.Println("Usage of ExtractMRIHeadInfo:")
	flag.PrintDefaults()
}
func main() {
	root := flag.String("root", "", "-root 数据存放路径")
	outputFileName := flag.String("o", "HeaderInfo.xlsx", "-o output.xlsx")
	strReg := flag.String("strReg", "nii.gz$", "-strReg nii.gz$")
	flag.Parse()
	if len(os.Args) <= 1 {
		displayHelp()
		return
	}
	strReg4Match := regexp.MustCompile(*strReg)
	var jsonFilesName []string
	_ = filepath.Walk(*root, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strReg4Match.MatchString(info.Name()) {
			_, errCommand := exec.Command("mrinfo", path, "-json_all", path+".json").CombinedOutput()
			if errCommand != nil {
				fmt.Println(errCommand)
			}
			jsonFilesName = append(jsonFilesName, path+".json")
		}
		return nil
	})

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
	res[0][0] = "json file"
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
