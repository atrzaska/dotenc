package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	if len(os.Args) < 3 {
		panic("Please specify wich environment should be used")
	}

	return os.Args[2]
}

func readEnvFile() []string {
	envPath := getEnvFilePath()
	envData, err := ioutil.ReadFile(envPath)
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

func readPublicKey() string {
	lines := readEnvFile()
	header := lines[0]

	if !strings.HasPrefix(header, PUBLIC_KEY_PREFIX) {
		panic("Public key comment not found at top of the env file ")
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

func writeFile(buffer bytes.Buffer) {
	envPath := getEnvFilePath()
	outFile, err := os.Create(envPath)
	check(err)
	defer outFile.Close()
	outFile.Write([]byte(buffer.String()))
}

func isParsable(line string) bool {
	return strings.Contains(line, "=") && !strings.HasPrefix(line, "#")
}

func isNotLastLine(i int, lines []string) bool {
	return i < len(lines)-1
}

func getEnvFilePath() string {
	env := readEnv()
	return fmt.Sprintf(".env.%s", env)
}

func getExecCommand() string {
	if len(os.Args) < 4 {
		panic("Please specify command to run")
	}

	return strings.Join(os.Args[3:], " ")
}

func createKeypair() crypto.Keypair {
	pubkey := readPublicKey()
	keyMap := readKeyMap()
	privkey := keyMap[pubkey]

	keyPair := crypto.Keypair{
		Public:  convertHexToBytes(pubkey),
		Private: convertHexToBytes(privkey),
	}

	return keyPair
}

func createEncrypter() *crypto.Encrypter {
	pubkey := readPublicKey()
	var keyPair crypto.Keypair
	err := keyPair.Generate()
	check(err)

	return keyPair.Encrypter(convertHexToBytes(pubkey))
}

func createDecrypter() *crypto.Decrypter {
	keyPair := createKeypair()
	return keyPair.Decrypter()
}

func decryptEnvToString() string {
	var buffer bytes.Buffer

	lines := readEnvFile()
	decrypter := createDecrypter()

	for i, line := range lines {
		var result string

		if isParsable(line) {
			key, value := splitEnvLine(line)
			decryptedValue, err := decrypter.Decrypt([]byte(value))
			check(err)
			result = key + "=" + string(decryptedValue)
		} else {
			result = line
		}

		buffer.WriteString(result)

		if isNotLastLine(i, lines) {
			buffer.WriteString("\n")
		}
	}

	return buffer.String()
}

func decryptEnv() {
	result := decryptEnvToString()
	fmt.Print(result)
}

func encryptEnv() {
	var buffer bytes.Buffer

	lines := readEnvFile()
	encrypter := createEncrypter()

	for i, line := range lines {
		var result string

		if isParsable(line) {
			key, value := splitEnvLine(line)
			encryptedValue, err := encrypter.Encrypt([]byte(value))
			check(err)
			result = key + "=" + string(encryptedValue)
		} else {
			result = line
		}

		buffer.WriteString(result)

		if isNotLastLine(i, lines) {
			buffer.WriteString("\n")
		}
	}

	writeFile(buffer)
}

func generateKeyPair() {
	var kp crypto.Keypair

	err := kp.Generate()
	check(err)

	fmt.Println("Public key: " + kp.PublicString())
	fmt.Println("Private key: " + kp.PrivateString())
	fmt.Println()
	fmt.Println("Add this line on top of your dotfile:")
	fmt.Println("# public_key: " + kp.PublicString())
	fmt.Println()
	fmt.Println("Add this line to your .dotenc file:")
	fmt.Println(kp.PublicString() + ": " + kp.PrivateString())
	fmt.Println()
	fmt.Println("Remember to ignore .dotenc in your version control system! You can use following command:")
	fmt.Println("echo \".dotenc\" >> .gitignore")
}

func printHelp() {
	fmt.Println("Dotenc is a small library to manage encrypted secrets using asymetric encryption.")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  dotenc [command]")
	fmt.Println("")
	fmt.Println("Available Commands:")
	fmt.Println("  encrypt [env]          Encrypt given environment file .env.[env]")
	fmt.Println("  e [env]                Shortcut for encrypt")
	fmt.Println("  decrypt [env]          Decrypt given environment file .env.[env] and print to STDOUT")
	fmt.Println("  d [env]                Shortcut for decrypt")
	fmt.Println("  generate               Generate new public and private key")
	fmt.Println("  g                      Shortcut for generate")
	fmt.Println("  exec [env] [command]   Decrypt and load env variables from .env.[env] file and run program [command]")
}

func stripExport(key string) string {
	return strings.ReplaceAll(key, "export ", "")
}

func loadEnv() {
	decryptedEnv := decryptEnvToString()
	lines := strings.Split(decryptedEnv, "\n")

	for _, line := range lines {
		if !isParsable(line) {
			continue
		}

		key, value := splitEnvLine(line)
		key = stripExport(key)
		os.Setenv(key, value)
	}
}

func execCommand() int {
	command := getExecCommand()
	loadEnv()
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}

	return 0
}

func readOperation() string {
	if len(os.Args) == 1 {
		return "help"
	}

	operation := os.Args[1]

	switch operation {
	case "exec":
		return "exec"
	case "generate":
		return "generate"
	case "g":
		return "generate"
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

func main() {
	operation := readOperation()

	switch operation {
	case "help":
		printHelp()
	case "generate":
		generateKeyPair()
	case "decrypt":
		decryptEnv()
	case "encrypt":
		encryptEnv()
	case "exec":
		os.Exit(execCommand())
	}
}
