package main

import (
	"fmt"
	"testing"
)

func TestGenClient(t *testing.T) {
	school := &SchoolInfo{
		Area:   "YO",
		Name:   "YO",
		Id:     "YO",
		UserId: "YO",
		UserPw: "YOYOYO",
	}
	fmt.Println(school.Name)
	client, err := createClient(*school)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(client.School.Name)
}
