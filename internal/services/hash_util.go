package services

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
)

type HashAlgorithm string

const (
	HashAlgorithmSHA256 HashAlgorithm = "sha256"
	HashAlgorithmSHA384 HashAlgorithm = "sha384"
	HashAlgorithmSHA512 HashAlgorithm = "sha512"
)

type HashUtil struct{}

func NewHashUtil() *HashUtil {
	return &HashUtil{}
}

func (h *HashUtil) SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func (h *HashUtil) SHA256Hex(data []byte) string {
	return hex.EncodeToString(h.SHA256(data))
}

func (h *HashUtil) SHA256Base64(data []byte) string {
	return base64.StdEncoding.EncodeToString(h.SHA256(data))
}

func (h *HashUtil) SHA512(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

func (h *HashUtil) SHA512Hex(data []byte) string {
	return hex.EncodeToString(h.SHA512(data))
}

func (h *HashUtil) SHA512Base64(data []byte) string {
	return base64.StdEncoding.EncodeToString(h.SHA512(data))
}

func (h *HashUtil) SHA384(data []byte) []byte {
	hash := sha512.Sum384(data)
	return hash[:]
}

func (h *HashUtil) SHA384Hex(data []byte) string {
	return hex.EncodeToString(h.SHA384(data))
}

func (h *HashUtil) SHA384Base64(data []byte) string {
	return base64.StdEncoding.EncodeToString(h.SHA384(data))
}

func SHA256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func SHA256HashHex(data []byte) string {
	return hex.EncodeToString(SHA256Hash(data))
}

func SHA256HashBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(SHA256Hash(data))
}

func SHA512Hash(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

func SHA512HashHex(data []byte) string {
	return hex.EncodeToString(SHA512Hash(data))
}

func SHA512HashBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(SHA512Hash(data))
}

func SHA384Hash(data []byte) []byte {
	hash := sha512.Sum384(data)
	return hash[:]
}

func SHA384HashHex(data []byte) string {
	return hex.EncodeToString(SHA384Hash(data))
}

func SHA384HashBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(SHA384Hash(data))
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64Decode(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

func Base64URLDecode(encoded string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(encoded)
}

func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

func HexDecode(encoded string) ([]byte, error) {
	return hex.DecodeString(encoded)
}

func ComputeDocumentHash(data []byte, algorithm HashAlgorithm) []byte {
	switch algorithm {
	case HashAlgorithmSHA256:
		return SHA256Hash(data)
	case HashAlgorithmSHA384:
		return SHA384Hash(data)
	case HashAlgorithmSHA512:
		return SHA512Hash(data)
	default:
		return SHA256Hash(data)
	}
}

func ComputeDocumentHashHex(data []byte, algorithm HashAlgorithm) string {
	return hex.EncodeToString(ComputeDocumentHash(data, algorithm))
}

func ComputeDocumentHashBase64(data []byte, algorithm HashAlgorithm) string {
	return base64.StdEncoding.EncodeToString(ComputeDocumentHash(data, algorithm))
}

func VerifyDocumentHash(data []byte, expectedHash []byte, algorithm HashAlgorithm) bool {
	computedHash := ComputeDocumentHash(data, algorithm)
	if len(computedHash) != len(expectedHash) {
		return false
	}
	for i := 0; i < len(computedHash); i++ {
		if computedHash[i] != expectedHash[i] {
			return false
		}
	}
	return true
}

func VerifyDocumentHashHex(data []byte, expectedHashHex string, algorithm HashAlgorithm) bool {
	expectedHash, err := hex.DecodeString(expectedHashHex)
	if err != nil {
		return false
	}
	return VerifyDocumentHash(data, expectedHash, algorithm)
}

func VerifyDocumentHashBase64(data []byte, expectedHashBase64 string, algorithm HashAlgorithm) bool {
	expectedHash, err := base64.StdEncoding.DecodeString(expectedHashBase64)
	if err != nil {
		return false
	}
	return VerifyDocumentHash(data, expectedHash, algorithm)
}

func DoubleHashSHA256(data []byte) []byte {
	firstHash := sha256.Sum256(data)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:]
}

func DoubleHashSHA512(data []byte) []byte {
	firstHash := sha512.Sum512(data)
	secondHash := sha512.Sum512(firstHash[:])
	return secondHash[:]
}

type HashResult struct {
	SHA256     string `json:"sha256"`
	SHA384     string `json:"sha384"`
	SHA512     string `json:"sha512"`
	SHA256Hex  string `json:"sha256_hex"`
	SHA384Hex  string `json:"sha384_hex"`
	SHA512Hex  string `json:"sha512_hex"`
	Base64     string `json:"base64"`
}

func ComputeAllHashes(data []byte) *HashResult {
	return &HashResult{
		SHA256:    SHA256HashBase64(data),
		SHA384:    SHA384HashBase64(data),
		SHA512:    SHA512HashBase64(data),
		SHA256Hex: SHA256HashHex(data),
		SHA384Hex: SHA384HashHex(data),
		SHA512Hex: SHA512HashHex(data),
		Base64:    Base64Encode(data),
	}
}