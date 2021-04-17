package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Shopify/ejson/crypto"
)

const KEY_MAP_FILE = ".dotenc"
const PUBLIC_KEY_PREFIX = "# public_key: "
const KEY_MAP_SEPARATOR = ": "
const PUBLIC_KEY_SIZE = 32 * 2

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readKeyMap() map[string]string {
	data, err := ioutil.ReadFile(KEY_MAP_FILE)
	dataContent := string(data)
	check(err)

	results := make(map[string]string)
	lines := strings.Split(dataContent, "\n")

	for _, line := range lines {
		isValid := strings.Contains(line, KEY_MAP_SEPARATOR)

		if !isValid {
			continue
		}

		split := strings.Split(line, KEY_MAP_SEPARATOR)
		public := split[0]
		private := split[1]
		results[public] = private
	}

	return results
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

func readPublicKey(path string) string {
	lines := readEnvFile(path)
	header := lines[0]

	if !strings.HasPrefix(header, PUBLIC_KEY_PREFIX) {
		panic("Public key comment not found at top of the env file " + path)
	}

	publicKey := strings.Replace(header, PUBLIC_KEY_PREFIX, "", 1)

	if len(publicKey) != PUBLIC_KEY_SIZE {
		panic("Public key has wrong length, expected 32 bytes")
	}

	return publicKey
}

func convertHexToBytes(hexString string) [32]byte {
	decodedString, err := hex.DecodeString(hexString)
	check(err)

	var out [32]byte
	copy(out[:], decodedString[:32])
	return out
}

func writeFile(writePath string, buffer bytes.Buffer) {
	outFile, err := os.Create(writePath)
	check(err)
	defer outFile.Close()
	outFile.Write([]byte(buffer.String()))
}

func createKeypair(readPath string) crypto.Keypair {
	pubkey := readPublicKey(readPath)
	keyMap := readKeyMap()
	privkey := keyMap[pubkey]

	myKP := crypto.Keypair{
		Public:  convertHexToBytes(pubkey),
		Private: convertHexToBytes(privkey),
	}

	return myKP
}

func createEncrypter(readPath string) *crypto.Encrypter {
	pubkey := readPublicKey(readPath)
	myKP := createKeypair(readPath)

	return myKP.Encrypter(convertHexToBytes(pubkey))
}

func createDecrypter(readPath string) *crypto.Decrypter {
	myKP := createKeypair(readPath)
	return myKP.Decrypter()
}

func isParsable(line string) bool {
	return strings.Contains(line, "=") && !strings.HasPrefix(line, "#")
}

func isNotLastLine(i int, lines []string) bool {
	return i < len(lines)-1
}

func decryptEnv() {
	env := readEnv()
	readPath := fmt.Sprintf(".env.%s.enc", env)
	writePath := fmt.Sprintf(".env.%s", env)
	lines := readEnvFile(readPath)
	var buffer bytes.Buffer
	decrypter := createDecrypter(readPath)

	for i, line := range lines {
		var outLine string

		if isParsable(line) {
			key, value := splitEnvLine(line)
			decryptedValue, err := decrypter.Decrypt([]byte(value))
			check(err)
			outLine = key + "=" + string(decryptedValue)
		} else {
			outLine = line
		}

		buffer.WriteString(outLine)

		if isNotLastLine(i, lines) {
			buffer.WriteString("\n")
		}
	}

	writeFile(writePath, buffer)
}

func encryptEnv() {
	env := readEnv()
	readPath := fmt.Sprintf(".env.%s", env)
	writePath := fmt.Sprintf(".env.%s.enc", env)
	lines := readEnvFile(readPath)
	var buffer bytes.Buffer
	encrypter := createEncrypter(readPath)

	for i, line := range lines {
		var outLine string

		if isParsable(line) {
			key, value := splitEnvLine(line)
			encryptedValue, err := encrypter.Encrypt([]byte(value))
			check(err)
			outLine = key + "=" + string(encryptedValue)
		} else {
			outLine = line
		}

		buffer.WriteString(outLine)

		if isNotLastLine(i, lines) {
			buffer.WriteString("\n")
		}
	}

	writeFile(writePath, buffer)
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
