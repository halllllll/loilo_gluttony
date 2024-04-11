package loilo

import (
	"fmt"
)

var (
	Host    = "https://n.loilo.tv"
	Entry   = fmt.Sprintf("%s/users/sign_in", Host)
	Home    = fmt.Sprintf("%s/dashboard", Host)
	classes = func() string { return fmt.Sprintf("%s/user_groups", Host) }
	// graduates    = fmt.Sprintf("%s/students?graduated=1", host)
	studentsXlsx = func(id int) string {
		return fmt.Sprintf("%s/console/api/students/export_xlsx?scope=current&school_id=%d&by_query=", Host, id)
	}
	teachersXlsx = func(id int) string {
		return fmt.Sprintf("%s/console/api/teachers/export_xlsx?scope=current&school_id=%d&by_query=", Host, id)
	}
	class = func(id int) string { return fmt.Sprintf("%s/user_groups/%d/memberships", Host, id) }
)

type SchoolInfo struct {
	Name             string
	InternalSchoolId int
}

type StudentInfo struct {
	Name      string
	Kana      string
	UserId    string
	GoogleSSO string
	MSSSO     string
}

type ClassInfo struct {
	Name     string
	Grade    string
	Code     int
	Start    string
	End      string
	GroupID  int
	Students []StudentInfo
}

func StudentListSheetHeader() []interface{} {
	return []interface{}{
		"ユーザーID", "氏名", "ふりがな", "パスワード", "Googleのメールアドレス", "Microsoftのメールアドレス", "学年", "クラス名",
	}
}

func TeacherListSheetHeader() []interface{} {
	return []interface{}{
		"ユーザーID", "氏名", "ふりがな", "パスワード", "Googleのメールアドレス", "Microsoftのメールアドレス",
	}
}

func GenStudentExelUrl(internalSchoolId int) string {
	return studentsXlsx(internalSchoolId)
}

func GenTeacherExelUrl(internalSchoolId int) string {
	return teachersXlsx(internalSchoolId)
}

func GenClassListUrl() string {
	return classes()
}

func GenClassUrl(groupId int) string {
	return class(groupId)
}
