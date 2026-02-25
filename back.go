package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

type Client struct {
	name          string
	ServerUrl     string
	Http          *http.Client
	private_bytes [32]byte
	signing_bytes [32]byte
}

type KeyFile struct {
	Salt        string `json:"salt,omitempty"`
	Private_key string `json:"private_key"`
	Signing_key string `json:"signing_key"`
}

type Responce struct {
	Status string `json:"status"`
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
	temp, _ := scrypt.Key([]byte(password), salt_b, 32768, 8, 1, 32)
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
func Rgister(c *Client, name string) error {
	return nil
}

// просто пинг сервера. можно проверять доступность
func ping(c *Client) (string, error) {
	resp, err := c.Http.Post(c.ServerUrl+"/ping", "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result Responce
	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return "", err
	}

	return result.Status, nil
}

func register(c *Client) {
	//генерим
	//кидаем запрос на регу
}
