package setup

import (
	"embed"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/halllllll/loilo_gluttony/v2/storage"
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
	SaveDirRoot  string
	Storage      *storage.UnityExcel
}

func NewProject(s *storage.UnityExcel) *Project {
	return &Project{
		DataDirName:  dataDirName,
		DataFileName: dataFileName,
		LogFileName:  logFileName,
		Storage:      s,
	}
}

// ファイルの確認・中身の返却と保存用フォルダの作成
func (proj *Project) Hello(vd *embed.FS) (*[]LoginRecord, error) {
	target := filepath.ToSlash(filepath.Join(proj.DataDirName, proj.DataFileName))
	buf, err := vd.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("error read file %w", err)
	}
	var loginrecords []LoginRecord
	// 一気に読み込む
	if err := gocsv.UnmarshalBytes(bom.Clean(buf), &loginrecords); err != nil {
		return nil, fmt.Errorf("error read csv - %w", err)
	}

	saveTo, err := CreateDirectory(filepath.Join(ct()))
	if err != nil {
		return nil, fmt.Errorf("error create save dir - %w", err)
	}

	proj.SaveDirRoot = saveTo
	return &loginrecords, nil
}

func CreateDirectory(targetPath string) (string, error) {
	var fileNum = 1
	var fileName = targetPath
	for {
		if _, err := os.Stat(fileName); err != nil {
			err = os.Mkdir(fileName, os.ModePerm)
			if err != nil && !os.IsExist(err) {
				t := time.Duration(rand.Float64()) * time.Microsecond
				utils.ErrLog.Printf("error mkdir: %s\n -- rechalenge after '%s' (microsecond)...", t, err)
				time.Sleep(t)
				continue
			}
			break

		} else {
			fileNum += 1
			fileName = fmt.Sprintf("%s_%d", targetPath, fileNum)
			continue
		}
	}
	return fileName, nil
}
