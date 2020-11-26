package encrypt

import (
	"errors"
	"math/rand"
	"os"
	"time"
)

var (
	NoPermission = errors.New("the role permission denied")
	NoRole       = errors.New("role is invalid")
	NoPubKey     = errors.New("pub_key is not exist")
	NoMasterKey  = errors.New("master_key is not exist")
)

func writeFile(filename string, text string) error {
	textF, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer textF.Close()

	_, err = textF.WriteString(text)
	if err != nil {
		return err
	}

	return textF.Sync()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
