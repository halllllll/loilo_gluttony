package scrape

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gocolly/colly"
	"github.com/halllllll/loilo_gluttony/v2/loilo"
	"github.com/halllllll/loilo_gluttony/v2/setup"
)

var (
	uas = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3864.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:68.0) Gecko/20100101 Firefox/68.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:92.0) Gecko/20100101 Firefox/92.0",
	}
	ua = func() string { return uas[rand.Intn(len(uas))] }
)

type ScrapeAgent struct {
	Collector  *colly.Collector
	SchoolInfo *loilo.SchoolInfo
}

func Login(loginInfo *setup.LoginRecord) (*ScrapeAgent, error) {
	c := colly.NewCollector(
		colly.AllowedDomains("n.loilo.tv"),
		colly.UserAgent(ua()),
		colly.Async(true), // あとでスクレイピングするときに使う
	)
	// random delay
	c.Limit(&colly.LimitRule{
		RandomDelay: 3 * time.Second,
	})
	enter := &toEnterInfo{
		collector: c,
		record:    loginInfo,
	}
	agent, nil := enter.knock("管理者メニュー")

	return agent, nil
}

// WIP
// 全クラス情報を取得
// （所属しているアカウントは含まず、クラスのみ）
// -> loilo.ClassListProps
func (agent *ScrapeAgent) GenClassesInfo() error {
	var props loilo.ClassListProps
	var errMsg string
	c := *agent.Collector.Clone()
	groupId := agent.SchoolInfo.InternalSchoolId

	c.OnHTML("#app-props", func(e *colly.HTMLElement) {
		data := e.Attr("data-props")
		if err := json.Unmarshal([]byte(data), &props); err != nil {
			errMsg += fmt.Sprintf("unmarshall error:\n%s\n", err)
			return
		}
	})

	c.Wait()
	if err := c.Visit(loilo.GenClassUrl(groupId)); err != nil {
		return fmt.Errorf("error data props - %w\n%s", err, errMsg)
	}
	c.Wait()

	return nil
}

// WIP
// 各クラスごとのデータを取得
// -> loilo.ClassProps
func (agent *ScrapeAgent) GetClassInfoById(groupId int) error {
	var props loilo.ClassProps
	var errMsg string
	c := *agent.Collector.Clone()

	c.OnHTML("#app-props", func(e *colly.HTMLElement) {
		data := e.Attr("data-props")
		if err := json.Unmarshal([]byte(data), &props); err != nil {
			errMsg += fmt.Sprintf("unmarshall error:\n%s\n", err)
			return
		}
	})

	c.Wait()
	if err := c.Visit(loilo.GenClassUrl(groupId)); err != nil {
		return fmt.Errorf("error on GenClassInfoById - %w\n%s", err, errMsg)
	}
	c.Wait()

	return nil
}

// URL先のコンテンツを決められた形式（filePath）で保存
func (agent *ScrapeAgent) SaveContent(url, filePath string) error {
	var errMsg string
	c := *agent.Collector.Clone()

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			errMsg += fmt.Sprintf("cant access %s - statuscode: %d\n", url, r.StatusCode)
			return
		}
		if err := r.Save(filePath); err != nil {
			errMsg += fmt.Sprintf("error save content - %s", err)
			return
		}
	})
	if err := c.Visit(url); err != nil {
		return fmt.Errorf("failed get content %s - \n %s \n %w -", agent.SchoolInfo.Name, errMsg, err)
	}
	c.Wait()
	return nil
}

// this is only for IN-DEV func
// save html file
func (agent *ScrapeAgent) DownloadAsStaticHTML(saveDir string, url string) error {
	c := *agent.Collector.Clone()

	// this is a only sample
	c.OnHTML("#app-props", func(e *colly.HTMLElement) {
		ioutil.WriteFile(filepath.Join(saveDir, "response.html"), e.Response.Body, os.ModePerm)
	})

	if err := c.Visit(url); err != nil {
		panic(err)
	}
	c.Wait()
	return nil
}

// this is only for IN-DEV func
// parse (local) static html file
func (agent *ScrapeAgent) ParseStaticHTML(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("not exist file %s - %w", path, err)
	}
	// t := &http.Transport{}
	// t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))

	// c := *agent.Collector.Clone()
	// c.SetRequestTimeout(10 * time.Second)
	// c.WithTransport(t)

	// // here your code...
	// c.OnResponse(func(r *colly.Response) {
	// 	fmt.Println("yay!")
	// 	fmt.Println(string(r.Body))
	// })

	// fmt.Println("reading static html file...")
	// if err := c.Visit(filepath.Join("file://", path)); err != nil {
	// 	return fmt.Errorf("%w", err)
	// }
	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)
	port := "5963"
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
	c := *agent.Collector.Clone()

	if err := c.Visit("http://localhost:5963"); err != nil {
		panic(err)
	}

	c.Wait()
	return nil
}
