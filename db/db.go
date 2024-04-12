package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"time"

	"github.com/halllllll/loilo_gluttony/v2/setup"
	"github.com/halllllll/loilo_gluttony/v2/storage"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize/v2"
)

//go:embed schema.sql
var schema embed.FS

type DB struct {
	loiloDB *sql.DB
	proj    *setup.Project
}

func CreateDB(ctx context.Context, proj *setup.Project) (*DB, func(), error) {
	schema, err := schema.ReadFile("schema.sql")
	if err != nil {
		return nil, func() {}, err
	}
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/loilo.sqlite", proj.SaveDirRoot))
	if err != nil {
		return nil, func() {}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, func() { _ = db.Close() }, err
	}
	_, err = db.ExecContext(ctx, string(schema))
	if err != nil {
		return nil, func() { _ = db.Close() }, err
	}

	stmt := `INSERT INTO info(goon) VALUES (?)`
	_, err = db.ExecContext(ctx, stmt, true)
	if err != nil {
		return nil, func() { _ = db.Close() }, err
	}

	loilodb := &DB{
		proj:    proj,
		loiloDB: db,
	}
	return loilodb, func() { _ = db.Close() }, err
}

func (db *DB) SetStudentDB(ctx context.Context) error {
	stusheetpath := filepath.Join(db.proj.SaveDirRoot, storage.StudentWorkBookName)
	f, err := excelize.OpenFile(stusheetpath)
	if err != nil {
		return err
	}
	rows, err := f.GetRows("sheet")
	if err != nil {
		return err
	}

	tx, err := db.loiloDB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() error {
		err := tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}()

	for _, row := range rows {
		for len(row) < 9 {
			row = append(row, "")
		}
		stmt := `INSERT INTO students(school_name, loilo_user_id, name, kana, password, google_account_id, ms_account_id, grade, class) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err := tx.ExecContext(ctx, stmt, row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[7], row[8])
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) SetTeacherDB(ctx context.Context) error {
	stusheetpath := filepath.Join(db.proj.SaveDirRoot, storage.TeacherWorkBookName)
	f, err := excelize.OpenFile(stusheetpath)
	if err != nil {
		return err
	}
	rows, err := f.GetRows("sheet")
	if err != nil {
		return err
	}

	tx, err := db.loiloDB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() error {
		err := tx.Commit()
		if err != nil {
			return err
		}
		return nil
	}()

	for _, row := range rows {
		for len(row) < 7 {
			row = append(row, "")
		}
		stmt := `INSERT INTO teachers(school_name, loilo_user_id, name, kana, password, google_account_id, ms_account_id) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err := tx.ExecContext(ctx, stmt, row[0], row[1], row[2], row[3], row[4], row[5], row[6])
		if err != nil {
			return err
		}
	}

	return nil
}
