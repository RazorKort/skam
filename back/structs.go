// пусть отдельным файлом висят
package back

import (
	"net/http"
)

type Client struct {
	Id             int
	name           string
	ServerUrl      string
	Http           *http.Client
	private_bytes  [32]byte
	signing_bytes  [32]byte
	token          string
	Friends        []User
	FriendsById    map[int]int
	SelectedFriend *User
}

// я ебал это делать отдельными структурами
// да я намешал всё в кучу и запросы и ответы
// а хули мне ты сделаешь?
type Body struct {
	Status         string    `json:"status,omitempty"`
	Name           string    `json:"name,omitempty"`
	Public_key     string    `json:"public_key,omitempty"`
	Verify_key     string    `json:"verify_key,omitempty"`
	Token          string    `json:"token,omitempty"`
	Signed_message string    `json:"signed_message,omitempty"`
	Seed           string    `json:"seed,omitempty"`
	Signed_seed    string    `json:"signed_seed,omitempty"`
	Friend_id      int       `json:"friend_id,omitempty"`
	Target_id      int       `json:"target_id,omitempty"`
	Id             int       `json:"id,omitempty"`
	Detail         string    `json:"detail,omitempty"`
	Message        string    `json:"message,omitempty"`
	Messages       []Message `json:"messages,omitempty"`
	Users          []User    `json:"users,omitempty"`
	Friends        []User    `json:"friends,omitempty"`
	New_name       string    `json:"new_name,omitempty"`
}

type User struct {
	Id           int    `json:"id,omitempty"`
	Name         string `json:"nickname"`
	Public_key   string `json:"public_key,omitempty"`
	Verify_key   string `json:"verify_key,omitempty"`
	Public_bytes [32]byte
	Verify_bytes [32]byte
	Shared_key   [32]byte
	Messages     []Message
	Loaded       bool
}

type Message struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Sender_id   int    `json:"sender_id"`
	Receiver_id int    `json:"receiver_id"`
	Message     string `json:"message"`
	Created_at  int    `json:"created_at,omitempty"`
	Plaintext   string
}

type KeyFile struct {
	Salt        string `json:"salt,omitempty"`
	Private_key string `json:"private_key"`
	Signing_key string `json:"signing_key"`
}
