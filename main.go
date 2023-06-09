package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	color "github.com/fatih/color"
	"github.com/gen2brain/beeep"
	"github.com/halllllll/loilo_gluttony/v2/loilo"
	"github.com/halllllll/loilo_gluttony/v2/scrape"
	"github.com/halllllll/loilo_gluttony/v2/setup"
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
	img, _ := notifyImg.Open("notify.png")
	fileInfo, _ := img.Stat()
	beeep.Notify(title, message, fileInfo.Name())
}

func init() {
	// 最低限プログラムを走らせることができるか確認 and 準備
	proj = setup.NewProject()
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
	// TODO

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

			saveDir, err := setup.CreateDirectory(filepath.Join(proj.SaveDirRoot, agent.SchoolInfo.Name))
			if err != nil {
				errCh <- fmt.Errorf("failed create save dir for %s - %s", agent.SchoolInfo.Name, err)
				return
			}

			internalId := agent.SchoolInfo.InternalSchoolId

			// TODO: SHOULD BE MATOMO BRAIN
			var subWg sync.WaitGroup
			subWg.Add(2)
			{
				// ONE
				go func() {
					defer subWg.Done()
					studentFile := filepath.Join(saveDir, fmt.Sprintf("%s__students.xlsx", agent.SchoolInfo.Name))
					if err := agent.SaveContent(loilo.GenStudentExelUrl(internalId), studentFile); err != nil {
						errCh <- fmt.Errorf("failed saving STUDENT content on %s - %w", agent.SchoolInfo.Name, err)
						return
					}
				}()
				// TWO
				go func() {
					defer subWg.Done()
					teacherFile := filepath.Join(saveDir, fmt.Sprintf("%s__teacherss.xlsx", agent.SchoolInfo.Name))
					if err := agent.SaveContent(loilo.GenTeacherExelUrl(internalId), teacherFile); err != nil {
						errCh <- fmt.Errorf("failed saving TEACHER content on %s - %w", agent.SchoolInfo.Name, err)
						return
					}
				}()
			}
		}(record)

		// output classes props info (test)
		// agent.GenClassesInfo()
		// agent.GetClassInfoById()

	}

	wg.Wait()
	// ごちゃごちゃしているが、ログインできた率とできなかった情報を出しているだけ
	utils.InfoLog.Printf("login successed -  %d/%d (%f)", len(failedLoginRecords), len(*loginRecords),
		float64(float64(len(failedLoginRecords))/float64(len(*loginRecords))))
	if len(*loginRecords) != len(failedLoginRecords) {
		utils.InfoLog.Println(yellow.Sprintln("PLEASE CHECK CAN'T LOGIN SCHOOL INFORMATION"))
		for idx, r := range failedLoginRecords {
			utils.InfoLog.Println(yellow.Sprintf("%d --- %s(%s)", idx+1, r.SchoolName, r.SchoolId))
		}
	}
	DesktopNotify{}.ShowNotify("BIG LOVE", "OVER!!!!!!!")
	utils.StdLog.Println("FINISH! byebyeﾉｼ")
	bufio.NewScanner(os.Stdin).Scan()
	os.Exit(1)
}
