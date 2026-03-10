package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"crypto/rand"
	"net/http"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
)

type Client struct {
	id            int
	name          string
	ServerUrl     string
	Http          *http.Client
	private_bytes [32]byte
	signing_bytes [32]byte
	token         string
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

type Message struct {
	Name        string `json:"name"`
	Sender_id   int    `json:"sender_id"`
	Receiver_id int    `json:"receiver_id"`
	Message     string `json:"message"`
	Created_at  string `json:"created_at,omitempty"`
}

type User struct {
	Friend_id    int    `json:"friend_id,omitempty"`
	Id           int    `json:"id,omitempty"`
	Name         string `json:"nickname"`
	Public_key   string `json:"public_key,omitempty"`
	Verify_key   string `json:"verify_key,omitempty"`
	Public_bytes [32]byte
	Verify_bytes [32]byte
}

// Https post helper
func HttpPost(c *Client, path string, body Body) (*Body, error) {
	json_b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(json_b)

	resp, err := c.Http.Post(c.ServerUrl+path, "application/json", bodyReader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if result.Status != "ok" {
		return nil, fmt.Errorf("server error: %s", result.Detail)
	}
	//проверка только для одной функции. мде...
	if result.Status == "none" {
		return nil, fmt.Errorf("no messages")
	}
	return &result, nil
}

// конструктор. Возвращает ссылку и http клиент
func NewClient(url string) (*Client, error) {
	return &Client{
		ServerUrl: url,
		Http:      &http.Client{},
	}, nil
}

// регистрация
func Register(c *Client, name string, password string) error {
	c.name = name
	//генерим priv pub
	public_b, private_b, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	//генерим подпись
	verify_b, signing_b, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	//оказывается подпись размером в 64 байта...
	seed_b := [32]byte(signing_b.Seed())
	c.private_bytes = *private_b
	c.signing_bytes = seed_b
	//сразу пишем в файл
	err = EncryptKey(password, c)
	if err != nil {
		return err
	}

	pubB64 := base64.StdEncoding.EncodeToString(public_b[:])
	verifB64 := base64.StdEncoding.EncodeToString(verify_b)
	body := Body{
		Name:       name,
		Public_key: pubB64,
		Verify_key: verifB64,
	}
	result, err := HttpPost(c, "/register", body)
	if err != nil {
		return err
	}

	c.token = result.Token
	c.id = result.Id
	return nil
}

// ауууф, выкатываем со дворов
func Auth(c *Client) error {
	//вычисляем public key
	var public_b [32]byte
	curve25519.ScalarBaseMult(&public_b, &c.private_bytes)
	public_key := base64.StdEncoding.EncodeToString(public_b[:])
	body := Body{
		Public_key: public_key,
	}

	result, err := HttpPost(c, "/auth-request", body)
	if err != nil {
		return err
	}

	private_key := ed25519.NewKeyFromSeed(c.signing_bytes[:])
	seed := []byte(result.Seed)
	signature := ed25519.Sign(private_key, seed)
	seedB64 := base64.StdEncoding.EncodeToString(seed)
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	body = Body{
		Public_key:     public_key,
		Signed_message: seedB64,
		Signed_seed:    signatureB64,
	}

	result, err = HttpPost(c, "/auth-verify", body)
	if err != nil {
		return err
	}

	c.token = result.Token
	c.id = result.Id
	c.name = result.Name
	return nil
}

// получить список друзей массив структур user[]
func GetFriends(c *Client) ([]User, error) {
	body := Body{
		Token: c.token,
	}

	result, err := HttpPost(c, "/friends", body)
	if err != nil {
		return nil, err
	}

	for i := range result.Friends {
		public_bytes, err := DecodeKey(result.Friends[i].Public_key)
		if err != nil {
			return nil, err
		}
		result.Friends[i].Public_bytes = *public_bytes

		verify_bytes, err := DecodeKey(result.Friends[i].Verify_key)
		if err != nil {
			return nil, err
		}
		result.Friends[i].Verify_bytes = *verify_bytes
	}

	return result.Friends, nil
}

// добавление друга
func AddFriend(c *Client, friend_id int) error {
	body := Body{
		Token:     c.token,
		Friend_id: friend_id,
	}
	_, err := HttpPost(c, "/addfriend", body)
	if err != nil {
		return err
	}
	return nil
}

// удаление друга
func RemoveFriend(c *Client, target_id int) error {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	_, err := HttpPost(c, "/removefriend", body)
	if err != nil {
		return err
	}
	return nil
}

// удаление чата. сразу у всех
func RemoveChat(c *Client, target_id int) error {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	_, err := HttpPost(c, "/removechat", body)
	if err != nil {
		return err
	}
	return nil
}

// удаление профиля
func RemoveProfile(c *Client) error {
	body := Body{
		Token: c.token,
	}
	_, err := HttpPost(c, "/remove", body)
	if err != nil {
		return err
	}
	return nil
}

// получение публичного ключа
func GetPublicKey(c *Client, target_id int) (*[32]byte, error) {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	result, err := HttpPost(c, "/getpublic", body)
	if err != nil {
		return nil, err
	}
	public_b, err := DecodeKey(result.Public_key)
	if err != nil {
		return nil, err
	}
	return public_b, nil
}

// грузим все сообщения. надо бы ограничить...
func LoadMessages(c *Client, target_id int) (*[]Message, error) {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	result, err := HttpPost(c, "/messages", body)
	if err != nil {
		return nil, err
	}

	return &result.Messages, nil
}

// ну давай, скажи что непонятно здесь
func ChangeName(c *Client, new_name string) error {
	body := Body{
		Token:    c.token,
		New_name: new_name,
	}
	_, err := HttpPost(c, "/changename", body)
	if err != nil {
		return err
	}
	c.name = new_name
	return nil
}

// поиск юзеров не требует авторизации.... надо бы исправить
// да и возвращает всех сразу. надо бы тоже исправить
func SearchUser(c *Client, name string) (*[]User, error) {
	body := Body{
		Name: name,
	}

	result, err := HttpPost(c, "/search", body)
	if err != nil {
		return nil, err
	}

	return &result.Users, nil
}

// просто пинг сервера
func Ping(c *Client) (string, error) {
	body := Body{}
	result, err := HttpPost(c, "/messages", body)
	if err != nil {
		return "", err
	}
	return result.Status, nil
}
