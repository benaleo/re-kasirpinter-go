package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

// pkcs7Pad adds PKCS#7 padding to the data
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpad removes PKCS#7 padding from the data
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	if len(data)%blockSize != 0 {
		return nil, errors.New("data is not block-aligned")
	}
	padding := int(data[len(data)-1])
	if padding > blockSize || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}

// GetAESSecret retrieves the AES secret key from environment
// Returns the raw hex string
func GetAESSecret() (string, error) {
	secret := os.Getenv("AES_SECRET")
	if secret == "" {
		return "", errors.New("AES_SECRET environment variable not set")
	}
	return secret, nil
}

// GetAESKeyBytes converts the secret string to bytes
// Supports both hex-encoded 32-byte keys and string passphrases
func GetAESKeyBytes() ([]byte, error) {
	secret, err := GetAESSecret()
	if err != nil {
		return nil, err
	}

	// Try to decode as hex first
	key, err := hex.DecodeString(secret)
	if err == nil {
		// Valid hex string
		if len(key) == 32 {
			return key, nil // AES-256 key
		}
		// If hex but not 32 bytes, treat as passphrase
	}

	// Treat as passphrase - use SHA256 to derive 32-byte key
	hash := sha256.Sum256([]byte(secret))
	return hash[:], nil
}

// Encrypt encrypts plaintext using AES-256-CBC
// Returns base64 encoded ciphertext (IV + ciphertext)
func Encrypt(plaintext string) (string, error) {
	// Get key bytes from hex secret
	key, err := GetAESKeyBytes()
	if err != nil {
		return "", err
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Convert plaintext to bytes and add PKCS#7 padding
	plaintextBytes := []byte(plaintext)
	paddedPlaintext := pkcs7Pad(plaintextBytes, aes.BlockSize)

	// Create byte slice for IV + ciphertext
	// AES block size is 16 bytes for IV
	ciphertext := make([]byte, aes.BlockSize+len(paddedPlaintext))

	// Generate random IV
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %v", err)
	}

	// Encrypt using CBC mode
	stream := cipher.NewCBCEncrypter(block, iv)
	stream.CryptBlocks(ciphertext[aes.BlockSize:], paddedPlaintext)

	// Return base64 encoded ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64 encoded ciphertext using AES-256-CBC
// Returns original plaintext
// Supports both standard format (IV + ciphertext) and crypto-js format (Salted__ + salt + key + iv + ciphertext)
func Decrypt(ciphertextBase64 string) (string, error) {
	// Decode base64 ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	// Check if it's crypto-js format (starts with "Salted__")
	if len(ciphertext) >= 8 && string(ciphertext[:8]) == "Salted__" {
		return decryptCryptoJS(ciphertext)
	}

	// Otherwise use standard format
	return decryptStandard(ciphertext)
}

// decryptStandard decrypts standard AES-256-CBC format (IV + ciphertext)
func decryptStandard(ciphertext []byte) (string, error) {
	// Get key bytes from hex secret
	key, err := GetAESKeyBytes()
	if err != nil {
		return "", err
	}

	// Validate ciphertext length (must have at least IV)
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Extract IV
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Validate ciphertext length (must be multiple of block size)
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	// Decrypt using CBC mode
	stream := cipher.NewCBCDecrypter(block, iv)
	stream.CryptBlocks(ciphertext, ciphertext)

	// Remove PKCS#7 padding
	plaintext, err := pkcs7Unpad(ciphertext, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("failed to unpad: %v", err)
	}

	// Return plaintext as string
	return string(plaintext), nil
}

// decryptCryptoJS decrypts crypto-js OpenSSL-compatible format
// Format: "Salted__" + 8-byte salt + ciphertext
func decryptCryptoJS(ciphertext []byte) (string, error) {
	// Try multiple approaches
	// Approach 1: Use salt as IV, remaining as ciphertext
	keyBytes, err := GetAESKeyBytes()
	if err != nil {
		return "", err
	}

	// Validate minimum length (Salted__ + 8 bytes salt + 16 bytes ciphertext)
	if len(ciphertext) < 24 {
		return "", errors.New("crypto-js ciphertext too short")
	}

	// Skip "Salted__" prefix
	dataAfterSalted := ciphertext[8:]

	// Next 8 bytes are salt - use as IV
	if len(dataAfterSalted) < 8 {
		return "", errors.New("no salt found")
	}
	salt := dataAfterSalted[:8]
	actualCiphertext := dataAfterSalted[8:]

	// Pad salt to 16 bytes for IV (if needed)
	iv := make([]byte, aes.BlockSize)
	copy(iv, salt)

	// Create cipher block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Validate ciphertext length
	if len(actualCiphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	// Decrypt
	stream := cipher.NewCBCDecrypter(block, iv)
	stream.CryptBlocks(actualCiphertext, actualCiphertext)

	// Remove PKCS#7 padding
	plaintext, err := pkcs7Unpad(actualCiphertext, aes.BlockSize)
	if err != nil {
		// If this fails, try passphrase approach
		return decryptCryptoJSPassphrase(ciphertext)
	}

	return string(plaintext), nil
}

// decryptCryptoJSPassphrase tries passphrase-based key derivation
func decryptCryptoJSPassphrase(ciphertext []byte) (string, error) {
	secret, err := GetAESSecret()
	if err != nil {
		return "", err
	}

	// Validate minimum length
	if len(ciphertext) < 16 {
		return "", errors.New("crypto-js ciphertext too short")
	}

	// Extract salt
	salt := ciphertext[8:16]
	ciphertext = ciphertext[16:]

	// Derive key and IV
	key, iv := deriveKeyAndIV(secret, salt)

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Validate ciphertext length (must be multiple of block size)
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	// Decrypt using CBC mode
	stream := cipher.NewCBCDecrypter(block, iv)
	stream.CryptBlocks(ciphertext, ciphertext)

	// Remove PKCS#7 padding
	plaintext, err := pkcs7Unpad(ciphertext, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("failed to unpad: %v (raw decrypted: %x)", err, ciphertext)
	}

	// Return plaintext as string
	return string(plaintext), nil
}

// deriveKeyAndIV implements OpenSSL's EVP_BytesToKey for key derivation
// Used by crypto-js for AES encryption
func deriveKeyAndIV(password string, salt []byte) (key, iv []byte) {
	// crypto-js uses MD5 for key derivation
	// For AES-256, we need 32 bytes key + 16 bytes IV = 48 bytes total
	// We'll concatenate MD5 hashes until we have enough bytes

	// Convert password to bytes
	passwordBytes := []byte(password)
	combined := append(passwordBytes, salt...)

	var derived []byte
	hash := combined

	for len(derived) < 48 { // 32 bytes key + 16 bytes IV
		hash = md5Hash(hash)
		derived = append(derived, hash...)
	}

	key = derived[:32]  // First 32 bytes for key
	iv = derived[32:48] // Next 16 bytes for IV
	return
}

// md5Hash computes MD5 hash of the data
func md5Hash(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}
