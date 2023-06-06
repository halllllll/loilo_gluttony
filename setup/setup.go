package setup

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/halllllll/loilo_gluttony/v2/loilo"
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
func (proj *Project) Hello(vd *embed.FS) ([]loilo.SchoolInfo, error) {
	if _, err := os.Stat(filepath.Join(proj.DataDirName, proj.DataFileName)); err != nil {
		return nil, fmt.Errorf("file not found - %w", err)
	}
	buf, err := vd.ReadFile(filepath.Join(proj.DataDirName, proj.DataFileName))
	if err != nil {
		return nil, fmt.Errorf("error read file %w", err)
	}
	reader := bytes.NewReader(buf)
	f := csv.NewReader(bom.NewReader(reader))
	schools := make([]loilo.SchoolInfo, 0)
	var header []string
	for idx := 0; ; idx++ {
		record, err := f.Read()
		if idx == 0 {
			header = record
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil || len(record) != len(header) {
			return nil, fmt.Errorf("error read line - %w", err)
		}
		school := loilo.SchoolInfo{
			Name:     record[0],
			SchoolId: record[1],
			AdminId:  record[2],
			AdminPw:  record[3],
		}
		schools = append(schools, school)
		fmt.Printf("%v\n", record)
	}

	return schools, nil
}
