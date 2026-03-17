package back

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

// Https post helper
func (c *Client) HttpPost(path string, body Body) (*Body, error) {
	json_b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(json_b)

	resp, err := c.Http.Post("https://"+c.ServerUrl+path, "application/json", bodyReader)
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
// он сразу пишет в файл
func (c *Client) Register(name string, password string) error {
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
	result, err := c.HttpPost("/register", body)
	if err != nil {
		return err
	}

	c.token = result.Token
	c.id = result.Id
	return nil
}

// ауууф, выкатываем со дворов
func (c *Client) Auth() error {
	//вычисляем public key
	var public_b [32]byte
	curve25519.ScalarBaseMult(&public_b, &c.private_bytes)
	public_key := base64.StdEncoding.EncodeToString(public_b[:])
	body := Body{
		Public_key: public_key,
	}

	result, err := c.HttpPost("/auth-request", body)
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

	result, err = c.HttpPost("/auth-verify", body)
	if err != nil {
		return err
	}

	c.token = result.Token
	c.id = result.Id
	c.name = result.Name
	return nil
}

// получить список друзей массив структур user[]
func (c *Client) GetFriends() ([]User, error) {
	body := Body{
		Token: c.token,
	}

	result, err := c.HttpPost("/friends", body)
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
func (c *Client) AddFriend(friend_id int) error {
	body := Body{
		Token:     c.token,
		Friend_id: friend_id,
	}
	_, err := c.HttpPost("/addfriend", body)
	if err != nil {
		return err
	}
	return nil
}

// удаление друга
func (c *Client) RemoveFriend(target_id int) error {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	_, err := c.HttpPost("/removefriend", body)
	if err != nil {
		return err
	}
	return nil
}

// удаление чата. сразу у всех
func (c *Client) RemoveChat(target_id int) error {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	_, err := c.HttpPost("/removechat", body)
	if err != nil {
		return err
	}
	return nil
}

// удаление профиля
func (c *Client) RemoveProfile() error {
	body := Body{
		Token: c.token,
	}
	_, err := c.HttpPost("/remove", body)
	if err != nil {
		return err
	}
	return nil
}

// получение публичного ключа
func (c *Client) GetPublicKey(target_id int) (*[32]byte, error) {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	result, err := c.HttpPost("/getpublic", body)
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
func (c *Client) LoadMessages(target_id int) (*[]Message, error) {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	result, err := c.HttpPost("/messages", body)
	if err != nil {
		return nil, err
	}

	return &result.Messages, nil
}

// ну давай, скажи что непонятно здесь
func (c *Client) ChangeName(new_name string) error {
	body := Body{
		Token:    c.token,
		New_name: new_name,
	}
	_, err := c.HttpPost("/changename", body)
	if err != nil {
		return err
	}
	c.name = new_name
	return nil
}

// поиск юзеров не требует авторизации.... надо бы исправить
// да и возвращает всех сразу. надо бы тоже исправить
func (c *Client) SearchUser(name string) (*[]User, error) {
	body := Body{
		Name: name,
	}

	result, err := c.HttpPost("/search", body)
	if err != nil {
		return nil, err
	}

	return &result.Users, nil
}

// просто пинг сервера
func (c *Client) Ping() (string, error) {
	body := Body{}
	result, err := c.HttpPost("/messages", body)
	if err != nil {
		return "", err
	}
	return result.Status, nil
}
