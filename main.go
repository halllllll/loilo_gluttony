package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
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
	"github.com/gen2brain/beeep"
	"github.com/halllllll/loilo_gluttony/v2/loilo"
	"github.com/halllllll/loilo_gluttony/v2/scrape"
	"github.com/halllllll/loilo_gluttony/v2/setup"
	"github.com/halllllll/loilo_gluttony/v2/utils"
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
	graduates    = fmt.Sprintf("%s/students?graduated=1", host)
	studentsXlsx = fmt.Sprintf("%s/students.xlsx", host)
	teachersXlsx = fmt.Sprintf("%s/teachers.xlsx", host)
	red          = color.New(color.Bold, color.BgHiBlack, color.FgHiRed)
	yellow       = color.New(color.Bold, color.BgHiBlack, color.FgHiYellow)
	green        = color.New(color.Bold, color.BgHiBlack, color.FgHiGreen)
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
		fmt.Printf("ri: %d\n", ri)
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

func (loilo *LoiloClient) GetMembers(url string) (result [][]string, err error) {
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
					td.Find("div > img").Each(func(idx int, sel *goquery.Selection) {
						if c, _ := sel.Attr("src"); strings.Contains(c, "icon_google") {
							u.googleSSO, _ = sel.Attr("title")
						} else if strings.Contains(c, "icon_microsoft") {
							u.msSSO, _ = sel.Attr("title")
						}
					})
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

func (loilo *LoiloClient) GetGraduates(url string) (result [][]string, err error) {
	// 卒業生はエンドポイント/graduate=1&page=Xっていうクエリになる
	// 複数ページが無いときはprev/nextのナビすら無いのでそこの数字とか取得するのは悪手？（なかったら1ページだけ、ということにすればいいけど）
	// 卒業生ページの構造がほぼ在校生のものと同じだったのでGetMembers流用
	pageNum := 0
	var tmpResult [][]string
	for {
		res, e := loilo.GetContent(fmt.Sprintf("%s&page=%d", url, pageNum))
		if e != nil {
			err = e
			return
		}
		doc, e := goquery.NewDocumentFromReader(res.Body)
		if e != nil {
			err = e
			return
		}
		tbodyChildrenNum := doc.Find("tbody").First().Children().Length()
		if tbodyChildrenNum == 0 {
			// 空なのでもう誰もいない
			break
		}
		pageNum += 1
		temp, e := loilo.GetMembers(fmt.Sprintf("%s&page=%d", url, pageNum))
		if e != nil {
			err = e
			return
		}
		// めんどくさいのでまとめて追加
		tmpResult = append(tmpResult, temp...)

	}
	// まとめて追加した結果毎回要らんヘッダーとかなぜか空のやつも取れてしまうのでとりあえず重複削除
	encounter := map[[5]string]int{}
	// 空は最初から無視するのであらかじめいれておく
	encounter[[5]string{"", "", "", "", ""}] = 1
	for _, v := range tmpResult {
		var vv [5]string
		copy(vv[:], v)
		if _, ok := encounter[vv]; !ok {
			encounter[vv] = 1
			result = append(result, v)
		} else {
			encounter[vv]++
		}
	}
	encounter[[5]string{"", "", "", "", ""}] -= 1 // 律儀（いらんでしょ）
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
		_, _ = classWb.NewSheet(sheetName)
		sheet, err := classWb.NewStreamWriter(sheetName)
		if err != nil {
			return err
		}
		members, err := loilo.GetMembers(fmt.Sprintf("%s/%s/memberships", classes, groupId))
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
	if err = classWb.SaveAs(filepath.FromSlash(fmt.Sprintf("%s/%s__classes.xlsx", path, loilo.School.Name))); err != nil {
		return err
	}
	utils.InfoLog.Println(yellow.Sprintf("%s 登録生徒 : %d \n", loilo.School.Name, len(memberCount)))
	return
}

func (loilo *LoiloClient) CreateGraduatesXlsx(graduatesList [][]string) (err error) {
	classWb := excelize.NewFile()
	classWb.NewSheet("Graduates") // 現状使ってるやつ
	classWb.DeleteSheet("Sheet1") // Sheet1はexcelizeでデフォルトで作られる/Stream中にシート名を変えるとエラーになったのでここでやる
	sheet, err := classWb.NewStreamWriter("Graduates")
	if err != nil {
		return err
	}
	path := filepath.FromSlash(Directory + "/" + loilo.School.Name)
	for rIdx, listRow := range graduatesList {
		// setRowが[]interface{}型のみ受け付けるので、スライスをそのまま使うことはできないしコピーして移すこともできない
		// ので、愚直にループでひとつずついれる
		row := make([]interface{}, len(listRow))
		for idx, v := range listRow {
			row[idx] = v
		}
		cell, err := excelize.CoordinatesToCellName(1, rIdx+1)
		if err != nil {
			return err
		}
		if err = sheet.SetRow(cell, row); err != nil {
			utils.ErrLog.Println(red.Sprint(err))
		}
	}
	if err = sheet.Flush(); err != nil {
		return err
	}
	if err = classWb.SaveAs(filepath.FromSlash(fmt.Sprintf("%s/%s__graduates.xlsx", path, loilo.School.Name))); err != nil {
		return err
	}

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

//go:embed notify.png
var notifyImg embed.FS

type Notify interface {
	ShowNotify(title, message string)
}
type DesktopNotify struct{}

func (d DesktopNotify) ShowNotify(title, message string) {
	img, _ := notifyImg.Open("notify.png")
	fileInfo, _ := img.Stat()
	beeep.Notify(title, message, fileInfo.Name())
}

//go:embed info/*
var LoginInfo embed.FS

func init() {
	utils.LoggingSetting("love.log")
	// ほかにもファイルとかは先に読んでおいたほうがいいのではないかという気がする
}

func main() {

	DesktopNotify{}.ShowNotify("BIG LOVE", "START")
	proj := setup.NewProject()
	loginInfo, err := proj.Hello(&LoginInfo)
	if err != nil {
		utils.ErrLog.Println(red.Sprintln("CANT START PROJECT."))
		utils.ErrLog.Println(red.Sprintln((err)))
		bufio.NewScanner(os.Stdin).Scan()
		os.Exit(1)
	}
	utils.StdLog.Println("save folder: ", proj.SaveDirRoot)

	var wg2 sync.WaitGroup
	for _, info := range loginInfo {
		wg2.Add(1)
		go func(data setup.LoginRecord) {
			defer wg2.Done()
			agent, err := scrape.Login(&data, proj)
			if err != nil {
				utils.ErrLog.Println(red.Sprintf("[%s] - failed to login - %s", data.SchoolName, err))
				return
			}
			utils.StdLog.Println(green.Sprintf("%s - START\n", agent.SchoolInfo.Name))

			saveDir, err := setup.CreateDirectory(filepath.Join(proj.SaveDirRoot, agent.SchoolInfo.Name))
			if err != nil {
				utils.ErrLog.Println(red.Sprintf("failed create save dir for %s - %s\n", agent.SchoolInfo.Name, err))
				return
			}
			internalId := agent.SchoolInfo.InternalSchoolId
			studentFile := filepath.Join(saveDir, fmt.Sprintf("%s__students.xlsx", agent.SchoolInfo.Name))
			agent.SaveContent(loilo.GenStudentExelUrl(internalId), studentFile)
			teacherFile := filepath.Join(saveDir, fmt.Sprintf("%s__teacherss.xlsx", agent.SchoolInfo.Name))

			agent.SaveContent(loilo.GenTeacherExelUrl(internalId), teacherFile)
			// wg2.Done()
		}(info)

		// saveDir, err := setup.CreateDirectory(filepath.Join(proj.SaveDirRoot, agent.SchoolInfo.Name))
		// if err != nil {
		// 	utils.ErrLog.Printf("failed create save dir for %s - %s\n", agent.SchoolInfo.Name, err)
		// 	continue
		// }
		// studentFile := filepath.Join(saveDir, fmt.Sprintf("%s__students.xlsx", agent.SchoolInfo.Name))

		// agent.SaveContent(agent.SchoolInfo.GenStudentExelUrl(), studentFile)
		// teacherFile := filepath.Join(saveDir, fmt.Sprintf("%s__teacherss.xlsx", agent.SchoolInfo.Name))

		// agent.SaveContent(agent.SchoolInfo.GenTeacherExelUrl(), teacherFile)
		// classes (test)
		// agent.GenClassesInfo()
		// agent.GetClassInfoById()

	}
	wg2.Wait()
	utils.StdLog.Println("FINISH! byebyeﾉｼ")
	bufio.NewScanner(os.Stdin).Scan()
	os.Exit(1)

	// フォルダ名 なんでもいいけど日付にしてる
	ct := time.Now().Format("2006_01_02")
	Directory, err = CreateSaveDirectory(ct)
	if err != nil {
		utils.ErrLog.Println(red.Sprint(err))
	}
	utils.StdLog.Println("save folder: ", Directory)

	// 配布するときは埋め込むけどcsvファイルから読みこむ
	entries, err := LoginInfo.ReadDir("info")
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

		buf, err := LoginInfo.ReadFile("info/" + entry.Name())
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
				Name:   record[0],
				Id:     record[1],
				UserId: record[2],
				UserPw: record[3],
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				// defer DesktopNotify{}.ShowNotify(fmt.Sprintf("%s !!", school.Name), "学校単位では終わったよ～")
				defer DesktopNotify{}.ShowNotify(school.Name, "おわったよ～")

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

	// DesktopNotify{}.ShowNotify("BIG LOVE", "おつかれ～～！！ばいば～～い")

	utils.StdLog.Println("FINISH! byebyeﾉｼ")
	bufio.NewScanner(os.Stdin).Scan()
}

func createClient(school SchoolInfo) (client *LoiloClient, err error) {
	values := url.Values{}
	values.Add("user[school][code]", school.Id)
	values.Add("user[username]", school.UserId)
	values.Add("user[password]", school.UserPw)

	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return
	}
	client = &LoiloClient{
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
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", Ua())
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	// ログインできてるか確認
	resp, err := client.GetContent(home)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	if doc, err := goquery.NewDocumentFromReader(resp.Body); doc.Find("h1.d-flex").Text() == "" || err != nil {
		if err != nil {
			return client, err
		} else {
			return client, errors.New(fmt.Sprintf("can't login to [ %s ] ...\n", school.Name))
		}
	}
	return
}

func gig(school SchoolInfo) (err error) {
	utils.StdLog.Printf("let's gig... %s\n", school.Name)
	schoolDir := filepath.FromSlash(Directory + "/" + school.Name)
	err = os.Mkdir(schoolDir, os.ModePerm)
	if err != nil {
		return
	}

	// cookie jarを用意してログイン。clientを使い回す
	client, err := createClient(school)
	if err != nil {
		return
	}
	// htmlファイルを保存する場合(ローカルから読み取る時用)
	/*
		b, err := io.ReadAll(res.Body)
		if err != nil {
			utils.ErrLog.Fatalln(err)
		}

		// fmt.Println(string(b))
		out, err := os.Create(schoolDir + "/" + "hoge.html")
		if err != nil {
			utils.ErrLog.Fatalln(err)
		}
		defer out.Close()
		_, err = io.Copy(out, bytes.NewReader(b))
	*/

	// 生徒情報xlsx（中身はgetしてるだけ）
	err = client.SaveXlsxFile(studentsXlsx, fmt.Sprintf("%s__students", school.Name))
	if err != nil {
		return
	}
	// 先生情報xlsx（中身はgetしてるだけ）
	err = client.SaveXlsxFile(teachersXlsx, fmt.Sprintf("%s__teachers", school.Name))
	if err != nil {
		return
	}

	// クラス情報のexcelを仕様通りに作ってやる
	classList, err := client.GetClasses(classes)
	if err != nil {
		return
	}
	utils.InfoLog.Println(yellow.Sprintf("%s class num: %d\n", school.Name, len(classList)-1))
	// ここでクラスのworkbookおよびsheet作成
	err = client.CreateClassesXlsx(classList)
	if err != nil {
		return
	}

	// 卒業生
	graduatesList, err := client.GetGraduates(graduates)
	if err != nil {
		return
	}
	if len(graduatesList) == 0 {
		// こういうことがある（なぜかヘッダーが入ってない）
		graduatesList = append(graduatesList, []string{"氏名", "ふりがな", "ユーザーID", "Google連携アカウント", "Microsoft連携アカウント"})
	}
	utils.InfoLog.Println(yellow.Sprintf("%s graduates: %d\n", school.Name, len(graduatesList)-1))
	err = client.CreateGraduatesXlsx(graduatesList)
	return
}
