package main

import "fmt"

func main() {
	password := "123"
	path := "session.key"
	client, err := NewClient("https://skam.su:10000")
	if err != nil {
		print("hui")
	}
	LoadKeys(client, path, password)
	err = Auth(client)
	if err != nil {
		fmt.Println(err)
	}
	friends, err := GetFriends(client)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(friends)
}
