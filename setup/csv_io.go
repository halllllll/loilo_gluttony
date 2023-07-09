package setup

type LoginRecord struct {
	SchoolName string `csv:"学校名"`
	SchoolId   string `csv:"学校ID"`
	AdminId    string `csv:"ユーザーID"`
	AdminPw    string `csv:"パスワード"`
}
