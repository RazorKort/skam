package back

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"os"

	"crypto/rand"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

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
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}

	//ключ их соли + пароль
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return err
	}
	var secretKey [32]byte
	copy(secretKey[:], key)

	//nonce и шифрование
	var nonce [24]byte
	_, err = rand.Read(nonce[:])
	if err != nil {
		return err
	}

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

// проверяем путь. если есть просим пароль. если нет вызываем регу
func CheckPath(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

// Грузим ключи из файла
// сразу вызывает decrypt и decode
func (c *Client) LoadKeys(path string, password string) error {
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

// вычисляем shared key для симметричного шифрования
func (c *Client) ComputeShared(friend *User) error {
	var shared [32]byte
	box.Precompute(&shared, &friend.Public_bytes, &c.private_bytes)

	friend.Shared_key = shared

	return nil
}

// дешифровка сообщения. записывает в message.plaintext
func DecryptMessage(msg *Message, friend User) error {
	Chiphertext, err := base64.StdEncoding.DecodeString(msg.Message)
	if err != nil {
		return err
	}

	nonceSize := 24
	if len(Chiphertext) < nonceSize {
		return fmt.Errorf("Chiphertext is too short")
	}
	var nonce [24]byte
	copy(nonce[:], Chiphertext[:nonceSize])

	plaintext, ok := box.OpenAfterPrecomputation(nil, Chiphertext[nonceSize:], &nonce, &friend.Shared_key)
	if !ok {
		//какого хера все нормальные функции возвращают err
		//а ебаный box ок
		return fmt.Errorf("Smth went wrong")
	}

	msg.Plaintext = string(plaintext)
	return nil
}

// шифрует сообщение, пишет в message.message
func EncryptMessage(msg *Message, friend User) error {
	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		//ахуеть, ты как сюда попал????
		return err
	}

	ciphertext := box.SealAfterPrecomputation(nil, []byte(msg.Plaintext), &nonce, &friend.Shared_key)
	out := append(nonce[:], ciphertext...)
	text := base64.StdEncoding.EncodeToString(out)
	msg.Message = text

	return nil
}
