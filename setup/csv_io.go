package setup

type LoginRecord struct {
	SchoolName string `csv:"学校名"`
	SchoolId   string `csv:"学校ID"`
	AdminID    string `csv:"ユーザーID"`
	AdminPW    string `csv:"パスワード"`
}
