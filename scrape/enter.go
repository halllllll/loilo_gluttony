package scrape

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/halllllll/loilo_gluttony/v2/loilo"
	"github.com/halllllll/loilo_gluttony/v2/setup"
	"github.com/halllllll/loilo_gluttony/v2/utils"
)

type toEnterInfo struct {
	collector *colly.Collector
	record    *setup.LoginRecord
}

// 2023/06/09
// landMarkText := "管理者メニュー"
func (info *toEnterInfo) knock(landMarkText string) (*ScrapeAgent, error) {
	var success bool = false
	c := info.collector
	school := &loilo.SchoolInfo{
		Name: info.record.SchoolName,
	}

	//　GO LOGIN
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

	// ほんとに管理者画面？
	c.OnHTML(".text-muted", func(e *colly.HTMLElement) {
		if success {
			if !strings.Contains(e.Text, landMarkText) {
				utils.ErrLog.Printf("error not container %s on html\n", info.record.SchoolName)
				success = false
			}
		}
	})

	// login operation (for Exponential BackOff)
	login := func() error {
		err := c.Post(loilo.Entry, map[string]string{
			"user[school][code]": info.record.SchoolId,
			"user[username]":     info.record.AdminId,
			"user[password]":     info.record.AdminPw,
			"commit":             "ログイン",
		})
		if err != nil {
			return fmt.Errorf("login error - %w", err)
		}
		c.Wait()
		// let's visiting!
		c.Visit(loilo.Home)
		return nil
	}

	// Exponential BackOff
	if err := withRetry(login); err != nil {
		return nil, fmt.Errorf("!OVER LOGIN RETRY COUNT! - %w", err)
	} else {
		// TODO: ログまわりをまともにする
		// utils.InfoLog.Printf("%s login data and POST scheme is valid!\n", school.Name)
	}

	c.Wait()
	if !success {
		return nil, fmt.Errorf("can't login (login data is invalid, or, changed HTML arch, especially '%s' ) - ", landMarkText)
	}
	agent := &ScrapeAgent{
		Collector:  c.Clone(),
		SchoolInfo: school,
	}
	return agent, nil
}
