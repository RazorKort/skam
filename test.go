package main

import "fmt"

func main() {
	password := "***"
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

	err = ChangeName(client, "razor1.0")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(client.name)
	users, err := SearchUser(client, "M")
	fmt.Println(users)
}
