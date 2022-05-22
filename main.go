package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"loilo/utils"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	color "github.com/fatih/color"
	"github.com/spkg/bom"
	"github.com/xuri/excelize/v2"
)

var (
	err          error
	Directory    string
	ngChars      = []rune{'/', ':', '?', '*', '[', ']', ' '}
	host         = "https://n.loilo.tv"
	entry        = fmt.Sprintf("%s/users/sign_in", host)
	home         = fmt.Sprintf("%s/school_dashboard", host)
	classes      = fmt.Sprintf("%s/user_groups", host)
	studentsXlsx = fmt.Sprintf("%s/students.xlsx", host)
	teachersXlsx = fmt.Sprintf("%s/teachers.xlsx", host)
	red          = color.New(color.Bold, color.BgHiBlack, color.FgHiRed)
	yellow       = color.New(color.Bold, color.BgHiBlack, color.BgHiYellow)
	green        = color.New(color.Bold, color.BgHiBlack, color.BgHiGreen)
)

func Ua() (ua string) {
	uas := []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3864.0 Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:68.0) Gecko/20100101 Firefox/68.0", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:92.0) Gecko/20100101 Firefox/92.0"}
	ua = uas[rand.Intn(len(uas))]
	return
}

type LoiloClient struct {
	School SchoolInfo
	Loilo  http.Client
}

// http.client.Doを使ってるだけなので返り値の変数名はつけないことにする（正しい作法なのかは不明）
func (loilo *LoiloClient) Do(req *http.Request) (*http.Response, error) {
	return loilo.Loilo.Do(req)
}

func (loilo *LoiloClient) GetContent(url string) (res *http.Response, err error) {
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", Ua())

	res, err = loilo.Do(req)
	return
}

func (loilo *LoiloClient) GetClasses(url string) (result [][]string, err error) {
	res, err := loilo.GetContent(url)
	if err != nil {
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	// 大きさ決め打ち
	classesLen := doc.Find("tr").Length()
	result = make([][]string, classesLen)
	for i := 0; i < classesLen; i++ {
		result[i] = make([]string, 6)
	}
	doc.Find("tr").Each(func(ri int, tr *goquery.Selection) {
		row := make([]string, 6)
		if ri == 0 {
			tr.Find("th").Each(func(ci int, th *goquery.Selection) {
				// 最初はなんかチェックボックスが入る
				if ci != 0 {
					row[ci-1] = th.Text()
				}
			})
			row[len(row)-1] = "グループID"
		} else {
			tr.Find("td").Each(func(ci int, td *goquery.Selection) {
				// 最初はなんかチェックボックスが入る
				if ci != 0 {
					row[ci-1] = td.Text()
				}
			})
			groupId, _ := tr.Find("input").Attr("value")
			row[len(row)-1] = groupId
		}
		result[ri] = row
	})
	return
}

func (loilo *LoiloClient) GetClassMembers(url string) (result [][]string, err error) {
	res, err := loilo.GetContent(url)
	if err != nil {
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	// 大きさ決め打ち
	memberLen := doc.Find("tr").Length()
	result = make([][]string, memberLen)
	for i := 0; i < memberLen; i++ {
		result[i] = make([]string, 5)
	}
	// 構造体でやります 効率がいいのかは不明
	type userOnClass struct {
		userName     string
		userKanaName string
		userId       string
		googleSSO    string
		msSSO        string
	}

	doc.Find("tr").Each(func(ri int, tr *goquery.Selection) {
		row := make([]string, 5)
		if ri == 0 {
			// ヘッダーを作る
			tr.Find("th").Each(func(ci int, th *goquery.Selection) {
				// 最初はなんかチェックボックスが入るなど考慮してヘッダー作成
				if ci != 0 {
					row[ci-1] = th.Text()
				}
			})
			// ヘッダーに足りないやつを追加
			row[len(row)-2] = "Google連携アカウント"
			row[len(row)-1] = "MS連携アカウント"
		} else {
			u := &userOnClass{}
			// td
			tr.Find("td").Each(func(ci int, td *goquery.Selection) {
				// 最初はなんかチェックボックスが入る
				switch ci {
				case 0:
				case 1:
					u.userName = td.Find("a").First().Text()
				case 2:
					u.userKanaName = td.Text()
				case 3:
					u.userId = td.Text()
					u.googleSSO, _ = td.Find("div > i.fa-google").Attr("title")
					u.msSSO, _ = td.Find("div > i.fa-microsoft").Attr("title")
				default:
				}
			})
			// 構造体のプロパティをここでハードコーディングして合わせるのダルいけどあとでなんか考える
			row[0] = u.userName
			row[1] = u.userKanaName
			row[2] = u.userId
			row[3] = u.googleSSO
			row[4] = u.msSSO
		}
		result[ri] = row
	})
	return
}

func (loilo *LoiloClient) SaveXlsxFile(url, fileName string) (err error) {
	if filepath.Ext(url) != ".xlsx" {
		return errors.New("It doesn't seems to be excel file (only support `xlsx` extension)")
	}
	path := filepath.FromSlash(Directory + "/" + loilo.School.Name)

	// 空ファイルを作って
	// 取得してきたレスポンスをそこにコピー
	file, err := os.Create(filepath.FromSlash(fmt.Sprintf("%s/%s.xlsx", path, fileName)))
	defer file.Close()
	if err != nil {
		return
	}
	res, err := loilo.GetContent(url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	_, err = io.Copy(file, res.Body)
	return
}

func (loilo *LoiloClient) CreateClassesXlsx(allClasses [][]string) (err error) {

	classWb := excelize.NewFile()
	classWb.NewSheet("Sheet")     // 現状使ってるやつ
	classWb.DeleteSheet("Sheet1") // Sheet1はexcelizeでデフォルトで作られる/Stream中にシート名を変えるとエラーになったのでここでやる
	sheet, err := classWb.NewStreamWriter("Sheet")
	if err != nil {
		return err
	}
	path := filepath.FromSlash(Directory + "/" + loilo.School.Name)

	// クラス名（シート反映用）
	classNames := make([]string, len(allClasses))
	// groupID（シート反映&スクレイピング用）
	groupIds := make([]string, len(allClasses))
	// ヘッダー streamWriterだと[]interface{}しか喰わない
	header := make([]interface{}, len(allClasses[0]))
	// 決まっているので
	headers := []string{"クラス名", "学年", "クラス参加コード", "開始日", "終了日", "グループID"}
	for idx, h := range headers {
		header[idx] = h
	}
	for rIdx := 0; rIdx <= len(allClasses); rIdx++ {
		if rIdx == 0 {
			if err := sheet.SetRow("A1", header); err != nil {
				utils.ErrLog.Println(red.Sprint(err))
			}
			continue
		}
		row := make([]interface{}, len(allClasses[0]))
		for idx, v := range allClasses[rIdx-1] {
			row[idx] = v

			// クラス名とグループIDを保存
			// なぜかrIdx == 0の場合（ヘッダー）でも作られるのでここでも無視
			if rIdx != 0 && idx == 0 {
				classNames[rIdx-1] = v
			}
			if rIdx != 0 && idx == 5 {
				groupIds[rIdx-1] = v
			}
		}
		cell, _ := excelize.CoordinatesToCellName(1, rIdx)
		if err = sheet.SetRow(cell, row); err != nil {
			utils.ErrLog.Println(red.Sprint(err))
		}
	}
	memberCount := make(map[string]bool)
	// 各クラスの情報を別シートごとに作成
	for classIdx, className := range classNames {
		if classIdx == 0 {
			// ヘッダー無視
			continue
		}
		groupId := groupIds[classIdx]
		// excelのシート名に使えない記号を置換
		// 一気にやるやり方がわからんのでreplaceする
		for _, c := range ngChars {
			className = strings.Replace(
				className,
				string(c),
				string('_'),
				-1,
			)
		}
		sheetName := fmt.Sprintf("%s_%s", className, groupId)

		// 2シート目以降 各クラス情報（エントリーポイント /user_groups/{groupid}/membership
		_ = classWb.NewSheet(sheetName)
		sheet, err := classWb.NewStreamWriter(sheetName)
		if err != nil {
			return err
		}
		members, err := loilo.GetClassMembers(fmt.Sprintf("%s/%s/memberships", classes, groupId))
		if err != nil {
			return err
		}
		for rowID := 0; rowID < len(members); rowID++ {
			row := make([]interface{}, len(members[rowID]))
			for colID := 0; colID < len(row); colID++ {
				row[colID] = members[rowID][colID]
			}
			cell, _ := excelize.CoordinatesToCellName(1, rowID+1)
			if err := sheet.SetRow(cell, row); err != nil {
				utils.ErrLog.Println(red.Sprint(err))
			}
			// 3番目がユーザーIDなので重複しない
			if _, ok := memberCount[fmt.Sprint(row[2])]; !ok {
				memberCount[fmt.Sprint(row[2])] = true
			}
		}
		if err := sheet.Flush(); err != nil {
			return err
		}
	}
	if err = sheet.Flush(); err != nil {
		return err
	}
	if err = classWb.SaveAs(filepath.FromSlash(fmt.Sprintf("%s/%sclasses.xlsx", path, loilo.School.Name))); err != nil {
		return err
	}
	utils.InfoLog.Println(yellow.Sprintf("%s : 登録生徒 %d \n", loilo.School.Name, len(memberCount)))
	return
}

func CreateSaveDirectory(target string) (directory string, err error) {
	cd, err := os.Getwd()
	if err != nil {
		return
	}
	var fileNum = 1
	var fileName = target
	for {
		err = os.Mkdir(fileName, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			return
		}
		if err != nil {
			fileNum += 1
			fileName = fmt.Sprintf("%s_%d", target, fileNum)
			continue
		}
		break
	}
	directory = filepath.FromSlash(cd + "/" + fileName)
	return
}

type SchoolInfo struct {
	Area   string
	Name   string
	Id     string
	UserId string
	UserPw string
}

//go:embed idpw/*
var idpw embed.FS

func init() {
	utils.LoggingSetting("love.log")
}

func main() {

	/*Script kiddie avoidance (experimental distribution)*/
	now := time.Now()
	target := time.Date(2022, 6, 10, 0, 0, 0, 0, time.Local)
	if !now.Before(target) {
		utils.ErrLog.Println(red.Sprint("!! EXPIRED !!"))
		utils.ErrLog.Println(red.Sprint("使用期限が切れました"))
		utils.ErrLog.Println(red.Sprintf("expired time: %s", target.Format("2006/01/02 15:04:05")))
		bufio.NewScanner(os.Stdin).Scan()
		os.Exit(1)
	}
	/*End of Script kiddie avoidance (experimental distribution)*/

	// フォルダ名 なんでもいいけど日付にしてる
	ct := time.Now().Format("2006_01_02")
	Directory, err = CreateSaveDirectory(ct)
	if err != nil {
		utils.ErrLog.Println(red.Sprint(err))
	}
	utils.StdLog.Println("save folder: ", Directory)

	// 配布するときは埋め込むけどcsvファイルから読みこむ
	entries, err := idpw.ReadDir("idpw")
	if err != nil {
		utils.ErrLog.Println(red.Sprint(err))
	}
	var wg sync.WaitGroup

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".csv" {
			continue
		}
		// buf, err := idpw.ReadFile(filepath.FromSlash("idpw/" + entry.Name())) やめたほうがいいっぽい　ビルドしたやつをwindowsで起動するとフォルダが見えなくて落ちた

		buf, err := idpw.ReadFile("idpw/" + entry.Name())
		if err != nil {
			utils.ErrLog.Println(red.Sprint(err))
		}
		reader := bytes.NewReader(buf)
		f := csv.NewReader(bom.NewReader(reader))

		idx := 0

		for {
			record, err := f.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				// なぜかcsvの行を読めず
				utils.ErrLog.Println(red.Sprint(err))
			}
			// 最初はヘッダーとする
			if idx == 0 {
				idx++
				continue
			}
			school := &SchoolInfo{
				Area:   record[0],
				Name:   record[1],
				Id:     record[2],
				UserId: record[3],
				UserPw: record[4],
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer utils.StdLog.Println(green.Sprintf("%s Done!\n", school.Name))
				err = gig(*school)
				if err != nil {
					utils.ErrLog.Println(red.Sprint(err))
				}
			}()
			idx++
		}
	}
	wg.Wait()
	utils.StdLog.Println("FINISH! byebyeﾉｼ")
	bufio.NewScanner(os.Stdin).Scan()
}

func gig(school SchoolInfo) (err error) {
	utils.StdLog.Printf("let's gig... %s\n", school.Name)
	schoolDir := filepath.FromSlash(Directory + "/" + school.Name)
	err = os.Mkdir(schoolDir, os.ModePerm)
	if err != nil {
		return err
	}
	values := url.Values{}
	values.Add("user[school][code]", school.Id)
	values.Add("user[username]", school.UserId)
	values.Add("user[password]", school.UserPw)

	// cookie jarを用意してログイン。clientを使い回す
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return err
	}
	client := &LoiloClient{
		School: school,
		Loilo: http.Client{
			Jar: jar,
		},
	}

	// ログインを試す
	req, err := http.NewRequest(
		"POST",
		entry,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", Ua())
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// ログインできてるか確認
	res, err = client.GetContent(home)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	if doc, err := goquery.NewDocumentFromReader(res.Body); doc.Find("h1.d-flex").Text() == "" || err != nil {
		if err != nil {
			return err
		} else {
			return errors.New(fmt.Sprintf("can't login to [ %s ] ...\n", school.Name))
		}
	}
	// 生徒情報xlsx（中身はgetしてるだけ）
	err = client.SaveXlsxFile(studentsXlsx, fmt.Sprintf("%sstudents", school.Name))
	if err != nil {
		return err
	}
	// 先生情報xlsx（中身はgetしてるだけ）
	err = client.SaveXlsxFile(teachersXlsx, fmt.Sprintf("%steachers", school.Name))
	if err != nil {
		return err
	}

	// クラス情報のexcelを仕様通りに作ってやる
	classList, err := client.GetClasses(classes)
	if err != nil {
		return err
	}
	utils.InfoLog.Println(yellow.Sprintf("%s class num: %d\n", school.Name, len(classList)-1))
	// ここでクラスのworkbookおよびsheet作成
	err = client.CreateClassesXlsx(classList)
	return
}
