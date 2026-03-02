package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"net/http"
	"os"

	"crypto/rand"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
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

type KeyFile struct {
	Salt        string `json:"salt,omitempty"`
	Private_key string `json:"private_key"`
	Signing_key string `json:"signing_key"`
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
	public_bytes [32]byte
	verify_bytes [32]byte
}

// дешифрует байтики
func DecryptKey(password string, salt string, bytes string) (*[32]byte, error) {

	bytes_b, err := base64.StdEncoding.DecodeString(bytes)
	if err != nil {
		return nil, err
	}

	salt_b, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, err
	}

	var key [32]byte
	temp, err := scrypt.Key([]byte(password), salt_b, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	copy(key[:], temp)

	var nonce [24]byte
	copy(nonce[:], bytes_b[:24])
	chiper := bytes_b[24:]

	plaintext, ok := secretbox.Open(nil, chiper, &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("Decryption failed")
	}
	if len(plaintext) != 32 {
		return nil, fmt.Errorf("key != 32 bytes")
	}
	var key_b [32]byte
	copy(key_b[:], plaintext)
	return &key_b, nil
}

// шифрует байтики и пишет в файл
func EncryptKey(password string, c *Client) error {
	//генерим соль
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)

	//ключ их соли + пароль
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return err
	}
	var secretKey [32]byte
	copy(secretKey[:], key)

	//nonce и шифрование
	var nonce [24]byte
	_, _ = rand.Read(nonce[:])

	cipher_priv := secretbox.Seal(nil, c.private_bytes[:], &nonce, &secretKey)
	full_priv := append(nonce[:], cipher_priv...)

	//новый nonce
	_, _ = rand.Read(nonce[:])

	cipher_sign := secretbox.Seal(nil, c.signing_bytes[:], &nonce, &secretKey)
	full_sign := append(nonce[:], cipher_sign...)

	//манипуляции с записыванием ключей
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	privB64 := base64.StdEncoding.EncodeToString(full_priv)
	signB64 := base64.StdEncoding.EncodeToString(full_sign)

	kf := KeyFile{
		Salt:        saltB64,
		Private_key: privB64,
		Signing_key: signB64,
	}

	jsonBytes, err := json.MarshalIndent(kf, "", " ")
	if err != nil {
		return err
	}

	if err := os.WriteFile("session.key", jsonBytes, 0600); err != nil {
		return err
	}
	return nil
}

// декодит из base 64 и приводит в нормальные байты размером 32
func DecodeKey(key string) (*[32]byte, error) {
	key_b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	if len(key_b) != 32 {
		return nil, fmt.Errorf("Expected 32 bytes")
	}
	var final_byte [32]byte
	copy(final_byte[:], key_b)
	return &final_byte, nil
}

// конструктор. Возвращает ссылку и http клиент
func NewClient(url string) (*Client, error) {
	return &Client{
		ServerUrl: url,
		Http:      &http.Client{},
	}, nil
}

// проверяем путь. если есть просим пароль. если нет вызываем регу
func CheckPath(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

// Грузим ключи из файла
func LoadKeys(c *Client, path string, password string) error {
	jsonBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var kf KeyFile
	err = json.Unmarshal(jsonBytes, &kf)
	if err != nil {
		return err
	}

	if password != "" {

		priv_b, err := DecryptKey(password, kf.Salt, kf.Private_key)
		if err != nil {
			return err
		}
		c.private_bytes = *priv_b

		sign_b, err := DecryptKey(password, kf.Salt, kf.Signing_key)
		if err != nil {
			return err
		}
		c.signing_bytes = *sign_b
	} else {
		priv_b, err := DecodeKey(kf.Private_key)
		if err != nil {
			return err
		}
		c.private_bytes = *priv_b

		sign_b, err := DecodeKey(kf.Signing_key)
		if err != nil {
			return err
		}
		c.signing_bytes = *sign_b
	}

	return nil
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
	EncryptKey(password, c)

	pubB64 := base64.StdEncoding.EncodeToString(public_b[:])
	verifB64 := base64.StdEncoding.EncodeToString(verify_b)
	body := Body{
		Name:       name,
		Public_key: pubB64,
		Verify_key: verifB64,
	}

	json_b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(json_b)

	resp, err := c.Http.Post(c.ServerUrl+"/register", "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}

	if result.Status != "ok" {
		return fmt.Errorf("server error: %s", result.Detail)
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
	json_b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/auth-request", "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return fmt.Errorf("Server error: %s", result.Detail)
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
	json_b, err = json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader = bytes.NewReader(json_b)

	resp, err = c.Http.Post(c.ServerUrl+"/auth-verify", "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return fmt.Errorf("Server error: %s", result.Detail)
	}
	c.token = result.Token
	c.id = result.Id
	c.name = result.Name
	return nil
}

// получить список друзей массив структур
func GetFriends(c *Client) ([]User, error) {
	body := Body{
		Token: c.token,
	}
	json_b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(json_b)

	resp, err := c.Http.Post(c.ServerUrl+"/friends", "application/json", bodyReader)
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
		return nil, fmt.Errorf("Server error : %s", result.Detail)
	}

	return result.Friends, nil
}

// добавление друга
func AddFriend(c *Client, friend_id int) error {
	body := Body{
		Token:     c.token,
		Friend_id: friend_id,
	}
	json_b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/addfriend", "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		//58 friend already friend
		//404 ну тут понятно
		return fmt.Errorf("Server error %s", result.Detail)
	}
	return nil
}

// удаление друга
func RemoveFriend(c *Client, target_id int) error {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	json_b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/removefriend", "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return fmt.Errorf("Server error %s", result.Detail)
	}
	return nil
}

// удаление чата. сразу у всех
func RemoveChat(c *Client, target_id int) error {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	json_b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/removechat", "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return fmt.Errorf("Server error %s", result.Detail)
	}
	return nil
}

// удаление профиля
func RemoveProfile(c *Client) error {
	body := Body{
		Token: c.token,
	}
	json_b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/remove", "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return fmt.Errorf("Server error %s", result.Detail)
	}
	return nil
}

// получение публичного ключа
func GetPublicKey(c *Client, target_id int) (*[32]byte, error) {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	json_b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/getpublic", "application/json", bodyReader)
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
		return nil, fmt.Errorf("Server error %s", result.Detail)
	}
	public_b, err := DecodeKey(result.Public_key)
	if err != nil {
		return nil, err
	}
	return public_b, nil
}

// грузим сообщения. надо бы ограничить...
func LoadMessages(c *Client, target_id int) (*[]Message, error) {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	json_b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/messages", "application/json", bodyReader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Status == "none" {
		return nil, fmt.Errorf("no messages")
	}
	return &result.Messages, nil
}

// ну давай, скажи что непонятно здесь
func ChangeName(c *Client, new_name string) error {
	body := Body{
		Token:    c.token,
		New_name: new_name,
	}
	json_b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/changename", "application/json", bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return fmt.Errorf("Smth went wrong and idk why %s", result.Status)
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
	json_b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(json_b)
	resp, err := c.Http.Post(c.ServerUrl+"/search", "application/json", bodyReader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Status == "nothing" {
		return nil, fmt.Errorf("nothing")
	}
	return &result.Users, nil
}

// просто пинг сервера
func Ping(c *Client) (string, error) {
	resp, err := c.Http.Post(c.ServerUrl+"/ping", "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result Body
	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return "", err
	}

	return result.Status, nil
}
