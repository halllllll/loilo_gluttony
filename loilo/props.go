package loilo

import "time"

type ClassListProps struct {
	Messages struct {
	} `json:"messages"`
	ImagePath struct {
		Logo           string `json:"logo"`
		GoogleIcon     string `json:"googleIcon"`
		MicrosoftIcon  string `json:"microsoftIcon"`
		AlertIcon      string `json:"alertIcon"`
		BlueFolderIcon string `json:"blueFolderIcon"`
		NarationIcon   string `json:"narationIcon"`
		PlayIcon       string `json:"playIcon"`
	} `json:"imagePath"`
	Locale      string `json:"locale"`
	CurrentUser struct {
		ID              int    `json:"id"`
		DisplayName     string `json:"displayName"`
		DisplayUserName string `json:"displayUserName"`
		IsAdmin         bool   `json:"isAdmin"`
		IsDistrictAdmin bool   `json:"isDistrictAdmin"`
		SignInType      any    `json:"signInType"`
		School          struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		} `json:"school"`
	} `json:"currentUser"`
	School struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
		Country  string `json:"country"`
	} `json:"school"`
	Pagination struct {
		CurrentPage int `json:"currentPage"`
		TotalPage   int `json:"totalPage"`
	} `json:"pagination"`
	DefaultStartDate  string `json:"defaultStartDate"`
	DefaultEndDate    string `json:"defaultEndDate"`
	CandidateStudents []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"candidateStudents"`
	UserGroupsByYear []struct {
		Year       any `json:"year"`
		UserGroups []struct {
			ID             int       `json:"id"`
			Name           string    `json:"name"`
			GradeString    string    `json:"gradeString"`
			CodeIsDisabled bool      `json:"codeIsDisabled"`
			Code           string    `json:"code"`
			StartAt        time.Time `json:"startAt"`
			FinishAt       time.Time `json:"finishAt"`
			IsDeleted      bool      `json:"isDeleted"`
		} `json:"userGroups"`
	} `json:"userGroupsByYear"`
	UserGroupsTotal int  `json:"userGroupsTotal"`
	UseReact        bool `json:"useReact"`
}
