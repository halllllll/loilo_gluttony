# When U want to update...
2023-06-09

## setup
data, as **Project**

```go
// 保存用のファイルやパス、データ用ファイル名など
var (
	dataDirName  = "info"
	dataFileName = "data.csv"
	logFileName  = "love.log"
```

`CreateDirectory` is my favorite (I think, ahh... 😌 imo)

## info
place of csv file for login data. it's refered on `main.go`, with `embed.FS`, as `LoginInfo`.

```go
//go:embed info/*
LoginInfo embed.FS
```

## loilo
loilonote archtecture and api(s).

(there are taken from admin view behaviors of on-browser)

```go
func GenStudentExelUrl(internalSchoolId int) string {
	return studentsXlsx(internalSchoolId)
}

func GenTeacherExelUrl(internalSchoolId int) string {
	return teachersXlsx(internalSchoolId)
}
```
...🧐（feeling BAD smells）

## scrape
todo

## 🤜 MAIN 🤛
todo