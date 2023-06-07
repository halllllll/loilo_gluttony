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
	studentsXlsx = func(id int) string { return fmt.Sprintf("%s/schools/%d/students.xlsx", Host, id) }
	teachersXlsx = func(id int) string { return fmt.Sprintf("%s/schools/%d/teachers.xlsx", Host, id) }
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

func (si *SchoolInfo) GenClassArch() {

}

func (si *SchoolInfo) GenStudentExelUrl() string {
	return studentsXlsx(si.InternalSchoolId)
}
func (si *SchoolInfo) GenTeacherExelUrl() string {
	return teachersXlsx(si.InternalSchoolId)
}

func (si *SchoolInfo) GenClassURL() string {
	return classes()
}
