package scrape

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/halllllll/loilo_gluttony/v2/loilo"
	"github.com/halllllll/loilo_gluttony/v2/setup"
	"github.com/halllllll/loilo_gluttony/v2/utils"
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
	Project    *setup.Project
}

func Login(loginInfo *setup.LoginRecord, project *setup.Project) (*ScrapeAgent, error) {

	school := &loilo.SchoolInfo{
		Name: loginInfo.SchoolName,
	}

	c := colly.NewCollector(
		colly.AllowedDomains("n.loilo.tv"),
		colly.UserAgent(ua()),
		colly.Async(true), // あとでスクレイピングするときに使う
	)
	// random delay
	c.Limit(&colly.LimitRule{
		RandomDelay: 3 * time.Second,
	})

	success := false
	// redirect home
	c.OnResponse(func(r *colly.Response) {
		reqUrl := r.Request.URL
		pattern := regexp.MustCompile(`/schools/(\d+)/dashboard`)
		if pattern.MatchString(reqUrl.Path) && r.StatusCode == 200 {
			id, err := strconv.Atoi(pattern.FindStringSubmatch(reqUrl.Path)[1])
			if err != nil {
				success = false
				utils.ErrLog.Printf("internal school id convert error: %s\n", err)
				return
			}
			success = true
			school.InternalSchoolId = id
		}
	})

	c.OnHTML("li.dropdown-header:nth-child(1)", func(e *colly.HTMLElement) {
		if success {
			if !strings.Contains(e.Text, loginInfo.SchoolName) {
				utils.ErrLog.Printf("error not container %s on html\n", loginInfo.SchoolName)
				success = false
			}
		}
	})

	// login
	err := c.Post(loilo.Entry, map[string]string{
		"user[school][code]": loginInfo.SchoolId,
		"user[username]":     loginInfo.AdminId,
		"user[password]":     loginInfo.AdminPw,
		"commit":             "ログイン",
	})
	if err != nil {
		return nil, fmt.Errorf("login error - %w", err)
	}
	c.Wait()
	c.Visit(loilo.Home)

	c.Wait()

	if !success {
		return nil, fmt.Errorf("can't login (or, login data is invalid, ex: schoolname) - ")
	}
	agent := &ScrapeAgent{
		Collector:  c.Clone(),
		SchoolInfo: school,
		Project:    project,
	}
	fmt.Println(agent.Project.SaveDirRoot, agent.Project.DataDirName)
	return agent, nil
}

func (agent *ScrapeAgent) SaveAsExcel(content [][]string, filePath string) {

}

func (agent *ScrapeAgent) GenClassInfo() error {
	c := *agent.Collector.Clone()
	c.OnResponse(func(r *colly.Response) {
		fmt.Println(r.StatusCode, r.Request.URL)
		fmt.Println(string(r.Body))
	})

	c.OnHTML("#app-props", func(e *colly.HTMLElement) {
		fmt.Println("table~~~")
		fmt.Println(e.Attr("data-props"))
	})
	c.Wait()
	fmt.Printf("url~~ %s\n", agent.SchoolInfo.GenClassURL())
	c.Visit(agent.SchoolInfo.GenClassURL())
	c.Wait()
	return nil
}

func (agent *ScrapeAgent) SaveContent(url, filePath string) error {
	var success bool

	agent.Collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			utils.ErrLog.Printf("cant access %s - statuscode: %d\n", url, r.StatusCode)
			success = false
			return
		}
		if err := r.Save(filePath); err != nil {
			utils.ErrLog.Printf("error save content - %s", err)
			success = false
			return
		}
		success = true
	})
	agent.Collector.Visit(url)
	agent.Collector.Wait()
	if !success {
		return fmt.Errorf("failed get content %s on %s", agent.SchoolInfo.Name, url)
	}
	return nil
}
