package main

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	color "github.com/fatih/color"
	"github.com/gen2brain/beeep"
	"github.com/halllllll/loilo_gluttony/v2/db"
	"github.com/halllllll/loilo_gluttony/v2/loilo"
	"github.com/halllllll/loilo_gluttony/v2/scrape"
	"github.com/halllllll/loilo_gluttony/v2/setup"
	"github.com/halllllll/loilo_gluttony/v2/storage"
	"github.com/halllllll/loilo_gluttony/v2/utils"
)

var (
	// for excel sheet
	ngChars = []rune{'/', ':', '?', '*', '[', ']', ' '}

	red    = color.New(color.Bold, color.FgHiRed)
	yellow = color.New(color.Bold, color.FgHiYellow)
	green  = color.New(color.Bold, color.FgHiGreen)

	//go:embed info/*
	LoginInfo embed.FS
	//go:embed notify.png
	notifyImg embed.FS

	proj         *setup.Project
	loginRecords *[]setup.LoginRecord
)

type Notify interface {
	ShowNotify(title, message string)
}
type DesktopNotify struct{}

func (d DesktopNotify) ShowNotify(title, message string) {
	img, _ := notifyImg.ReadFile("notify.png")
	tmpFile, _ := os.CreateTemp("", "notify-*.png")
	defer tmpFile.Close()
	_, _ = tmpFile.Write(img)
	beeep.Notify(title, message, tmpFile.Name())
}

func init() {
	// 最低限プログラムを走らせることができるか確認 and 準備
	unityExcel := storage.NewUnityExcel()
	proj = setup.NewProject(unityExcel)

	utils.LoggingSetting(proj.LogFileName)
	if data, err := proj.Hello(&LoginInfo); err != nil {
		utils.ErrLog.Println(red.Sprintln("CANT START PROJECT."))
		utils.ErrLog.Println(red.Sprintln((err)))
		bufio.NewScanner(os.Stdin).Scan()
		os.Exit(1)
	} else {
		loginRecords = data
	}
	DesktopNotify{}.ShowNotify("BIG LOVE", "START")
	utils.StdLog.Println("save folder: ", proj.SaveDirRoot)
	// ほかにも必要なファイルとか構造体は先に生成したり参照できるようにしとくといい気がする
}

func main() {
	ctx := context.Background()
	
	var wg sync.WaitGroup
	failedLoginRecords := make([]setup.LoginRecord, 0)

	// 後続のgoroutine内のエラーハンドリング　全部ここ
	errCh := make(chan error)
	go func() {
		for err := range errCh {
			utils.ErrLog.Println(red.Sprintln(err))
		}
	}()

	// 並行　メニュー

	for _, record := range *loginRecords {
		wg.Add(1)
		go func(data setup.LoginRecord) {
			defer wg.Done()
			agent, err := scrape.Login(&data)
			if err != nil {
				errCh <- fmt.Errorf("[%s] - failed to login - %w", data.SchoolName, err)
				failedLoginRecords = append(failedLoginRecords, data)
				return
			}
			utils.StdLog.Println(green.Sprintf("%s - START", agent.SchoolInfo.Name))

			// download standard xlsx, export 「生徒一覧」 and 「先生一覧」
			basicExportExcel(agent, proj, errCh)

		}(record)

	}

	wg.Wait()
	// ごちゃごちゃしているが、ログインできた率とできなかった情報を出しているだけ
	utils.InfoLog.Printf("login successed -  %d/%d (%f)", len(*loginRecords)-len(failedLoginRecords), len(*loginRecords),
		float64(float64(len(*loginRecords)-len(failedLoginRecords))/float64(len(*loginRecords)))*100)
	if len(failedLoginRecords) > 0 {
		utils.InfoLog.Println(yellow.Sprintln("PLEASE CHECK CAN'T LOGIN SCHOOL INFORMATION"))
		for idx, r := range failedLoginRecords {
			utils.InfoLog.Println(yellow.Sprintf("%d --- %s(%s)", idx+1, r.SchoolName, r.SchoolId))
		}
	}

	// integrate each sheet
	// 正常終了はos.Exitするんでdeferは通用しない
	proj.Storage.DeleteDefaultSheet()
	proj.Storage.Flush()
	proj.Storage.Save(proj.SaveDirRoot)
	proj.Storage.Close()

	fmt.Print("with server mode? (Y/n) > ")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		in := strings.ToLower(scanner.Text())
		if in == "y" {
			// save as db
			loilodb, closefunc, err := db.CreateDB(ctx, proj)
			if err != nil {
				closefunc()
				utils.ErrLog.Println(fmt.Errorf("%w", err))
			}
			defer closefunc()
			if err := loilodb.SetStudentDB(ctx); err != nil {
				utils.ErrLog.Println(err)
			}
			if err := loilodb.SetTeacherDB(ctx); err != nil {
				utils.ErrLog.Println(err)
			}
		}
		break
	}

	DesktopNotify{}.ShowNotify("BIG LOVE", "OVER!!!!!!!")
	utils.StdLog.Println("FINISH! byebyeﾉｼ")
	bufio.NewScanner(os.Stdin).Scan()
	os.Exit(1)
}

func basicExportExcel(agent *scrape.ScrapeAgent, proj *setup.Project, errCh chan error) {
	internalId := agent.SchoolInfo.InternalSchoolId
	saveDir, err := setup.CreateDirectory(filepath.Join(proj.SaveDirRoot, agent.SchoolInfo.Name))
	if err != nil {
		errCh <- fmt.Errorf("failed create save dir for %s - %w", agent.SchoolInfo.Name, err)
		return
	}

	// ONE
	type studentExcelResult struct {
		err error
	}
	studentChan := make(chan studentExcelResult)
	defer close(studentChan)

	go func(ch chan<- studentExcelResult) {
		studentFile := filepath.Join(saveDir, fmt.Sprintf("%s__students.xlsx", agent.SchoolInfo.Name))
		err := agent.SaveContent(loilo.GenStudentExelUrl(internalId), studentFile)
		if err != nil {
			ch <- studentExcelResult{err: fmt.Errorf("failed save content - %w", err)}
			return
		}
		if err = proj.Storage.AppendSSW(studentFile, agent.SchoolInfo.Name); err != nil {
			errCh <- err
		}

		ch <- studentExcelResult{err}

	}(studentChan)

	// TWO
	type teacherExcelResult struct {
		saveFilePath string
		err          error
	}
	teacherChan := make(chan teacherExcelResult)
	defer close(teacherChan)

	go func(ch chan<- teacherExcelResult) {
		teacherFile := filepath.Join(saveDir, fmt.Sprintf("%s__teacherss.xlsx", agent.SchoolInfo.Name))
		err := agent.SaveContent(loilo.GenTeacherExelUrl(internalId), teacherFile)
		if err != nil {
			ch <- teacherExcelResult{saveFilePath: "", err: fmt.Errorf("failed error content - %w", err)}
			return
		}
		if err = proj.Storage.AppendTSW(teacherFile, agent.SchoolInfo.Name); err != nil {
			errCh <- err
		}

		ch <- teacherExcelResult{saveFilePath: teacherFile, err: err}

	}(teacherChan)

	for i := 0; i < 2; i++ {
		select {
		case stu := <-studentChan:
			if stu.err != nil {
				errCh <- stu.err
				break
			}
		case tch := <-teacherChan:
			if tch.err != nil {
				errCh <- tch.err
				break
			}
		}
	}
}
