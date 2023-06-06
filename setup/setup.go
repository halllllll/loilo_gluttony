package setup

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/halllllll/loilo_gluttony/v2/utils"
	"github.com/spkg/bom"
)

// 保存用のファイルやパス、データ用ファイル名など
type Project struct {
	DataDirName  string
	DataFileName string
	LogFileName  string
}

func NewProject() *Project {
	return new(Project)
}

// ファイルの確認と保存用フォルダの作成
func Hello(vd *embed.FS) bool {
	entries, err := vd.ReadDir("info")
	if err != nil {
		utils.ErrLog.Println(err)
		log.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() != "data.csv" {
			continue
		}
		// read data.csv
		buf, err := vd.ReadFile("info/data.csv")
		if err != nil {
			log.Fatal(err)
		}
		reader := bytes.NewReader(buf)
		f := csv.NewReader(bom.NewReader(reader))
		for {
			record, err := f.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%v\n", record)
		}
	}

	return false
}
