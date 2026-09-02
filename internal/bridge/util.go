package bridge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

func digestFor(path string, raw []byte) FileDigest {
	sum := sha256.Sum256(raw)
	return FileDigest{
		Path:   filepath.ToSlash(path),
		Bytes:  len(raw),
		Lines:  lineCount(raw),
		SHA256: "sha256:" + hex.EncodeToString(sum[:]),
	}
}

func DigestFor(path string, raw []byte) FileDigest {
	return digestFor(path, raw)
}

func lineCount(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	count := 1
	for _, value := range raw {
		if value == '\n' {
			count++
		}
	}
	if raw[len(raw)-1] == '\n' {
		count--
	}
	return count
}

func JSON(value interface{}) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}

func writeOutputFile(path string, raw []byte) error {
	return os.WriteFile(path, raw, 0o644)
}
