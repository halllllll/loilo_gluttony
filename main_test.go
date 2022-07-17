package main

import (
	"loilo/utils"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateGraduateXlsx(t *testing.T) {
	/*
		school := &SchoolInfo{
			Area:   "X",
			Name:   "X",
			Id:     "X",
			UserId: "X",
			UserPw: "X",
		}
		client, err := createClient(*school)
		if err != nil {
			t.Fatal(err)
		}
		if school.Name != client.School.Name {
			t.Fatal(err)
		}
		graduatesList, err := client.GetGraduates(graduates)
		if err != nil {
			t.Fatal(err)
		}
	*/
	ct := time.Now().Format("2006_01_02")
	Directory, err = CreateSaveDirectory(ct)
	if err != nil {
		utils.ErrLog.Println(red.Sprint(err))
	}
	utils.StdLog.Println("save folder: ", Directory)

	var graduateList = [][]string{
		[]string{"氏名", "ふりがな", "ユーザーID", "Google連携アカウント", "MS連携アカウント"},
		[]string{"次元大介", "じげん　だいすけ", "konbatto-magunamu@peace.love", "konbatto-magunamu@peace.love", "konbatto-magunamu@peace.love"},
		[]string{"矢澤にこ", "やざわ　にこ", "smile@pink.niko", "smile@pink.niko", "smile@pink.niko"},
	}
	school := &SchoolInfo{
		Area:   "地球",
		Name:   "地球防衛隊第三ラグランジュ観測所",
		Id:     "xxxxx",
		UserId: "xxxxx",
		UserPw: "xxxxx",
	}

	client := &LoiloClient{
		School: *school,
		Loilo:  http.Client{},
	}

	schoolDir := filepath.FromSlash(Directory + "/" + school.Name)
	err = os.Mkdir(schoolDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = client.CreateGraduatesXlsx(graduateList)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("わーい")
}
