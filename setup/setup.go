package setup

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spkg/bom"
)

// 保存用のファイルやパス、データ用ファイル名など
var (
	dataDirName  = "info"
	dataFileName = "data.csv"
	logFileName  = "love.log"
)

type Project struct {
	DataDirName  string
	DataFileName string
	LogFileName  string
}

func NewProject() *Project {
	return &Project{
		DataDirName:  dataDirName,
		DataFileName: dataFileName,
		LogFileName:  logFileName,
	}
}

// ファイルの確認と保存用フォルダの作成
func (proj *Project) Hello(vd *embed.FS) ([]SchoolInfo, error) {
	if _, err := os.Stat(filepath.Join(proj.DataDirName, proj.DataFileName)); err != nil {
		return fmt.Errorf("file not found - %w", err)
	}
	buf, err := vd.ReadFile(filepath.Join(proj.DataDirName, proj.DataFileName))
	if err != nil {
		return fmt.Errorf("error read file %w", err)
	}
	reader := bytes.NewReader(buf)
	f := csv.NewReader(bom.NewReader(reader))
	for idx := 0; ; idx++ {
		record, err := f.Read()
		if idx == 0 {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error read line - %w", err)
		}
		fmt.Printf("%v\n", record)
	}

	return nil
}
