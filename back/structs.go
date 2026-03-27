// пусть отдельным файлом висят
package back

import (
	"net/http"
)

type Client struct {
	Id             int
	Name           string
	ServerUrl      string
	Http           *http.Client
	private_bytes  [32]byte
	signing_bytes  [32]byte
	token          string
	Friends        []User
	FriendsById    map[int]int
	Find           []User
	SelectedFriend *User
	WS             *WSClient
	WsMsgChan      chan Message
	WsSendChan     chan MessageToSend
	Cnt            int
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
	Receiver_id    int       `json:"receiver_id,omitempty"`
	Created_at     int       `json:"created_at,omitempty"`
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
	Type        string `json:"type"`
	Id          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Sender_id   int    `json:"sender_id,omitempty"`
	Receiver_id int    `json:"receiver_id,omitempty"`
	Message     string `json:"message,omitempty"`
	Created_at  int    `json:"created_at,omitempty"`
	Tmp_id      int    `json:"tmp_id,omitempty"`
	Plaintext   string
	Sended      bool
}
type MessageToSend struct {
	Sender_id   int    `json:"sender_id"`
	Receiver_id int    `json:"receiver_id"`
	Message     string `json:"message"`
	Created_at  int    `json:"created_at"`
	Tmp_id      int    `json:"tmp_id,omitempty"`
}
type KeyFile struct {
	Salt        string `json:"salt,omitempty"`
	Private_key string `json:"private_key"`
	Signing_key string `json:"signing_key"`
}
