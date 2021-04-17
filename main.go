package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const PASSWORD_FILE = ".dotenc"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func readPassword() string {
	passwordData, err := ioutil.ReadFile(PASSWORD_FILE)
	check(err)
	password := string(passwordData)
	return strings.ReplaceAll(password, "\n", "")
}

func readEnv() string {
	return os.Args[2]
}

func readOperation() string {
	operation := os.Args[1]

	switch operation {
	case "decrypt":
		return "decrypt"
	case "d":
		return "decrypt"
	case "encrypt":
		return "encrypt"
	case "e":
		return "encrypt"

	default:
		panic("Invalid operation: " + operation)
	}
}

func readEnvFile(envFile string) []string {
	envData, err := ioutil.ReadFile(envFile)
	check(err)
	envContent := string(envData)

	return strings.Split(envContent, "\n")
}

func splitEnvLine(line string) (string, string) {
	splitResult := strings.SplitN(line, "=", 2)
	id := splitResult[0]
	value := splitResult[1]

	return id, value
}

func encryptValue(value string) string {
	password := readPassword()
	encryptedValue := encrypt([]byte(value), password)

	return hex.EncodeToString(encryptedValue)
}

func decryptValue(value string) string {
	password := readPassword()
	decodedValue, err := hex.DecodeString(value)
	check(err)
	decryptedValue := decrypt(decodedValue, password)

	return string(decryptedValue)
}

func decryptEnv() {
	env := readEnv()
	readPath := fmt.Sprintf(".env.%s.enc", env)
	writePath := fmt.Sprintf(".env.%s", env)
	lines := readEnvFile(readPath)
	outFile, err := os.Create(writePath)
	check(err)

	defer outFile.Close()

	for _, line := range lines {
		isValid := strings.Contains(line, "=")

		if !isValid {
			continue
		}

		id, value := splitEnvLine(line)

		decryptedValueString := decryptValue(value)

		outLine := id + "=" + decryptedValueString + "\n"

		outFile.Write([]byte(outLine))
	}
}

func encryptEnv() {
	env := readEnv()
	readPath := fmt.Sprintf(".env.%s", env)
	writePath := fmt.Sprintf(".env.%s.enc", env)
	lines := readEnvFile(readPath)
	outFile, err := os.Create(writePath)
	check(err)

	defer outFile.Close()

	for _, line := range lines {
		isValid := strings.Contains(line, "=")

		if !isValid {
			continue
		}

		id, value := splitEnvLine(line)

		encryptedValueString := encryptValue(value)

		outLine := id + "=" + encryptedValueString + "\n"

		outFile.Write([]byte(outLine))
	}
}

func main() {
	operation := readOperation()

	switch operation {
	case "decrypt":
		decryptEnv()
	case "encrypt":
		encryptEnv()
	}
}
