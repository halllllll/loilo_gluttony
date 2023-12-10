package storage

import (
	"path/filepath"
	"sync"

	"github.com/halllllll/loilo_gluttony/v2/loilo"
	"github.com/halllllll/loilo_gluttony/v2/utils"
	"github.com/xuri/excelize/v2"
)

type UnityExcel struct {
	studentUnityExcel    *excelize.File
	teacherUnityExcel    *excelize.File
	studentUnityExcelSW  *excelize.StreamWriter
	teacherUnityExcelSW  *excelize.StreamWriter
	curStudentSheetSWRow int
	curTeacherSheetSWRow int
	mu                   sync.Mutex
}

// DLしてきたexcelのデフォルト
var sheetName = "sheet"

// excelのデフォルト（破壊）
var defaultSheetName = "Sheet1"

var studentWorkBookName = "student_all.xlsx"
var teacherWorkBookName = "teacher_all.xlsx"

func NewUnityExcel() *UnityExcel {
	var studentUnityExcel = excelize.NewFile()
	var teacherUnityExcel = excelize.NewFile()

	_, err := studentUnityExcel.NewSheet(sheetName)
	if err != nil {
		utils.ErrLog.Fatal(err)
	}

	_, err = teacherUnityExcel.NewSheet(sheetName)
	if err != nil {
		utils.ErrLog.Fatal(err)
	}

	ssw, err := studentUnityExcel.NewStreamWriter(sheetName)
	if err != nil {
		utils.ErrLog.Fatal(err)
	}
	tsw, err := teacherUnityExcel.NewStreamWriter(sheetName)
	if err != nil {
		utils.ErrLog.Fatal(err)
	}

	// prepared header
	// なぜかmakeの第二引数で呼ぶと不正なExcelファイルになったので
	var length = len(loilo.StudentListSheetHeader()) + 1
	vals := make([]interface{}, length)
	vals[0] = "学校名"
	for i, v := range loilo.StudentListSheetHeader() {
		vals[i+1] = v
	}
	ssw.SetRow("A1", vals)

	length = len(loilo.TeacherListSheetHeader()) + 1
	vals = make([]interface{}, length)
	vals[0] = "学校名"
	for i, v := range loilo.TeacherListSheetHeader() {
		vals[i+1] = v
	}
	tsw.SetRow("A1", vals)

	unityExcel := &UnityExcel{
		studentUnityExcel:    studentUnityExcel,
		teacherUnityExcel:    teacherUnityExcel,
		studentUnityExcelSW:  ssw,
		teacherUnityExcelSW:  tsw,
		curStudentSheetSWRow: 1, // 1 order and ignore header
		curTeacherSheetSWRow: 1, // 1 order and ignore header
	}

	return unityExcel
}

func (s *UnityExcel) DeleteDefaultSheet() {
	if err := s.studentUnityExcel.DeleteSheet(defaultSheetName); err != nil {
		utils.ErrLog.Fatal(err)
	}

	if err := s.teacherUnityExcel.DeleteSheet(defaultSheetName); err != nil {
		utils.ErrLog.Fatal(err)
	}
}

func (s *UnityExcel) Save(path string) {

	if err := s.studentUnityExcel.SaveAs(filepath.Join(path, studentWorkBookName)); err != nil {
		utils.ErrLog.Fatal(err)
	}

	if err := s.teacherUnityExcel.SaveAs(filepath.Join(path, teacherWorkBookName)); err != nil {
		utils.ErrLog.Fatal(err)
	}

}

func (s *UnityExcel) Flush() {
	if err := s.studentUnityExcelSW.Flush(); err != nil {
		utils.ErrLog.Fatal(err)
	}
	if err := s.teacherUnityExcelSW.Flush(); err != nil {
		utils.ErrLog.Fatal(err)
	}
}

func (s *UnityExcel) Close() {
	if err := s.studentUnityExcel.Close(); err != nil {
		utils.ErrLog.Fatal(err)
	}
	if err := s.teacherUnityExcel.Close(); err != nil {
		utils.ErrLog.Fatal(err)
	}
}

// for student
func (s *UnityExcel) AppendSSW(filePath string, schoolName string) error {

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	rows, err := f.Rows(sheetName)
	if err != nil {
		return err
	}
	for rowIdx := 0; rows.Next(); rowIdx++ {
		row, err := rows.Columns()
		if err != nil {
			return err
		}
		if rowIdx == 0 {
			// spoil header
			continue
		}
		val := make([]interface{}, len(row)+1)
		val[0] = schoolName
		for i, v := range row {
			val[i+1] = v
		}
		s.mu.Lock()
		s.curStudentSheetSWRow++
		cell, _ := excelize.CoordinatesToCellName(1, s.curStudentSheetSWRow)
		s.mu.Unlock()

		if err := s.studentUnityExcelSW.SetRow(cell, val); err != nil {
			return err
		}
	}

	return nil
}

// for student
func (s *UnityExcel) AppendTSW(filePath string, schoolName string) error {

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	rows, err := f.Rows(sheetName)
	if err != nil {
		return err
	}
	for rowIdx := 0; rows.Next(); rowIdx++ {
		row, err := rows.Columns()
		if err != nil {
			return err
		}
		if rowIdx == 0 {
			// spoil header
			continue
		}
		val := make([]interface{}, len(row)+1)
		val[0] = schoolName
		for i, v := range row {
			val[i+1] = v
		}
		s.mu.Lock()
		s.curTeacherSheetSWRow++
		cell, _ := excelize.CoordinatesToCellName(1, s.curTeacherSheetSWRow)
		s.mu.Unlock()

		if err := s.teacherUnityExcelSW.SetRow(cell, val); err != nil {
			return err
		}
	}

	return nil
}
