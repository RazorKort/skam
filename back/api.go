package back

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

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
	return &result, nil
}

// конструктор. Возвращает ссылку и http клиент
func NewClient(url string) (*Client, error) {
	return &Client{
		ServerUrl: url,
		Http:      &http.Client{},
		Cnt:       1,
	}, nil
}

// регистрация
// он сразу пишет в файл
func (c *Client) Register(name string, password string) error {
	c.Name = name
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
	err = c.EncryptKey(password)
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
	c.Id = result.Id
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
	c.Id = result.Id
	c.Name = result.Name
	return nil
}

// получить список друзей, добавить в client
func (c *Client) GetFriends() error {
	body := Body{Token: c.token}

	result, err := c.HttpPost("/friends", body)
	if err != nil {
		return err
	}

	// Инициализация если нужно
	if c.FriendsById == nil {
		c.FriendsById = make(map[int]int)
	}

	// Для каждого нового друга
	for i := range result.Friends {
		// Декодируем ключи
		public_bytes, err := DecodeKey(result.Friends[i].Public_key)
		if err != nil {
			return err
		}
		result.Friends[i].Public_bytes = *public_bytes

		verify_bytes, err := DecodeKey(result.Friends[i].Verify_key)
		if err != nil {
			return err
		}
		result.Friends[i].Verify_bytes = *verify_bytes

		shared := c.ComputeShared(result.Friends[i])
		result.Friends[i].Shared_key = shared

		// Проверяем, есть ли уже такой друг
		if idx, exists := c.FriendsById[result.Friends[i].Id]; exists {
			// обновляем существующего друга по индексу
			c.Friends[idx].Name = result.Friends[i].Name
			c.Friends[idx].Public_key = result.Friends[i].Public_key
			c.Friends[idx].Verify_key = result.Friends[i].Verify_key
			c.Friends[idx].Public_bytes = result.Friends[i].Public_bytes
			c.Friends[idx].Verify_bytes = result.Friends[i].Verify_bytes
			c.Friends[idx].Shared_key = result.Friends[i].Shared_key

		} else {
			// добавляем нового
			c.Friends = append(c.Friends, result.Friends[i])
			c.FriendsById[result.Friends[i].Id] = len(c.Friends) - 1
		}
	}

	return nil
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
	err = c.GetFriends()
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
// когда я сделаю так, чтобы грузилась только часть сообщений...
func (c *Client) LoadMessages(target_id int) error {
	body := Body{
		Token:     c.token,
		Target_id: target_id,
	}
	result, err := c.HttpPost("/messages", body)
	if err != nil {
		return err
	}

	idx, ok := c.FriendsById[target_id]
	if !ok {
		return fmt.Errorf("Friend not found")
	}

	// Создаем карту существующих ID

	//когда буду грузить частями, здаесь тоже проходиться по слайсу
	existingMessages := make(map[int]bool)
	//например n = 50 to n+50
	for i, msg := range c.Friends[idx].Messages {
		if c.Friends[idx].Messages[i].Id != -1 {
			existingMessages[msg.Id] = true
		}
	}

	// Добавляем новые сообщения
	for i := range result.Messages {
		if !existingMessages[result.Messages[i].Id] {
			result.Messages[i].Sended = true
			c.Friends[idx].Messages = append(c.Friends[idx].Messages, result.Messages[i])
		}
	}

	// Сортируем сообщения по timestamp
	//от меньшего к большему, дегенерат ебучий
	// not sended внизу
	sort.SliceStable(c.Friends[idx].Messages, func(i, j int) bool {
		if c.Friends[idx].Messages[i].Sended != c.Friends[idx].Messages[j].Sended {
			return c.Friends[idx].Messages[i].Sended
		}
		return c.Friends[idx].Messages[i].Created_at < c.Friends[idx].Messages[j].Created_at
	})

	return nil
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
	c.Name = new_name
	return nil
}

// поиск юзеров не требует авторизации.... надо бы исправить
// да и возвращает всех сразу. надо бы тоже исправить
func (c *Client) SearchUser(name string) error {
	body := Body{
		Name: name,
	}

	result, err := c.HttpPost("/search", body)
	if err != nil {
		return err
	}
	c.Find = result.Users
	return nil
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

func (c *Client) AddMessage(text string) (*Message, error) {
	c.Cnt++
	msg := Message{
		Id:          c.Cnt,
		Plaintext:   text,
		Sender_id:   c.Id,
		Receiver_id: c.SelectedFriend.Id,
		Created_at:  int(time.Now().UnixMilli()),
		Sended:      false,
	}
	err := EncryptMessage(&msg, *c.SelectedFriend)
	if err != nil {
		return nil, err
	}
	// Добавляем локально (оптимистичное обновление)
	if c.SelectedFriend != nil {
		c.SelectedFriend.Messages = append(c.SelectedFriend.Messages, msg)
		return &msg, nil
	}
	return nil, fmt.Errorf("No selected friends")
}

// отправляем сообщение
// falback на http не такая и плохая идея..., но надо придумать как обновлять такие сообщения
func (c *Client) SendMessage(msg Message) error {

	// Отправляем через WebSocket если есть
	if c.WS != nil {
		return c.WS.SendMessage(msg)
	}

	// Fallback на HTTP
	body := Body{
		Token:       c.token,
		Receiver_id: c.SelectedFriend.Id,
		Created_at:  int(time.Now().UnixMilli()),
	}
	_, err := c.HttpPost("/send", body)
	return err
}
