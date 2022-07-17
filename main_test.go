package main

import (
	"fmt"
	"testing"
)

func TestGenClient(t *testing.T) {
	school := &SchoolInfo{
		Area:   "X",
		Name:   "X",
		Id:     "X",
		UserId: "X",
		UserPw: "X",
	}
	fmt.Println(school.Name)
	client, err := createClient(*school)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(client.School.Name)
	graduatesList, err := client.GetGraduates(graduates)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for i, v := range graduatesList {
		fmt.Println(i, v)
		count += 1
	}
	fmt.Printf("合計で %d 個あった\n", count)
}
