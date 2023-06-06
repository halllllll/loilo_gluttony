package setup

import (
	"embed"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/halllllll/loilo_gluttony/v2/utils"
	"github.com/spkg/bom"
)

// 保存用のファイルやパス、データ用ファイル名など
var (
	dataDirName  = "info"
	dataFileName = "data.csv"
	logFileName  = "love.log"
	ct           = func() string { return time.Now().Format("2006_01_02") }
)

type Project struct {
	DataDirName  string
	DataFileName string
	LogFileName  string
	SaveDirName  string
}

func NewProject() *Project {
	return &Project{
		DataDirName:  dataDirName,
		DataFileName: dataFileName,
		LogFileName:  logFileName,
	}
}

// ファイルの確認・中身の返却と保存用フォルダの作成
func (proj *Project) Hello(vd *embed.FS) ([]LoginRecord, error) {
	if _, err := os.Stat(filepath.Join(proj.DataDirName, proj.DataFileName)); err != nil {
		return nil, fmt.Errorf("file not found - %w", err)
	}
	buf, err := vd.ReadFile(filepath.Join(proj.DataDirName, proj.DataFileName))
	if err != nil {
		return nil, fmt.Errorf("error read file %w", err)
	}
	var schools []LoginRecord
	if err := gocsv.UnmarshalBytes(bom.Clean(buf), &schools); err != nil {
		return nil, fmt.Errorf("error read csv - %w", err)
	}
	saveTo, err := createSaveDirectory(ct())
	if err != nil {
		return nil, fmt.Errorf("error create save dir - %w", err)
	}
	fmt.Printf("created? %s\n", saveTo)
	proj.SaveDirName = saveTo
	return schools, nil
}

func createSaveDirectory(target string) (string, error) {
	cd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error GetWd - %w ", err)
	}
	var fileNum = 1
	var fileName = target
	for {
		if _, err := os.Stat(fileName); err != nil {
			err = os.Mkdir(fileName, os.ModePerm)
			if err != nil && !os.IsExist(err) {
				t := time.Duration(rand.Int63n(1)) * time.Microsecond
				utils.ErrLog.Printf("error mkdir: %s\n -- rechalenge after '%s' (microsecond)...", t, err)
				time.Sleep(t)
				continue
			}
			break

		} else {
			fileNum += 1
			fileName = fmt.Sprintf("%s_%d", target, fileNum)
			continue
		}
	}
	directory := filepath.FromSlash(cd + "/" + fileName)
	return directory, nil
}
