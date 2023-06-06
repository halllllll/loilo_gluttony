package loilo

import (
	"fmt"
)

var (
	Host  = "https://n.loilo.tv"
	Entry = fmt.Sprintf("%s/users/sign_in", Host)
	Home  = fmt.Sprintf("%s/dashboard", Host)
	// classes      = fmt.Sprintf("%s/user_groups", host)
	// graduates    = fmt.Sprintf("%s/students?graduated=1", host)
	// studentsXlsx = fmt.Sprintf("%s/students.xlsx", host)
	// teachersXlsx = fmt.Sprintf("%s/teachers.xlsx", host)
)

type SchoolInfo struct {
	Name             string
	Id               string
	AdminId          string
	AdminPw          string
	InternalSchoolId int
}

func (si *SchoolInfo) GetClassURL() string {
	return fmt.Sprintf("%s/%d/classes", Host, si.InternalSchoolId)
}
