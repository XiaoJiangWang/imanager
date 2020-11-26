package encrypt

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/golang/glog"

	"imanager/pkg/config"
)

const (
	OpServiceRole = "op_service"
	AdminRole     = "admin"
	UserRole      = "user"
)

var roleMap = map[string]string{
	OpServiceRole: "1",
	AdminRole:     "2",
	UserRole:      "3",
}

var (
	MasterKeyFileName = "master_key"
	PubKeyFileName    = "pub_key"
)

func init() {
	encryptDir := config.GetConfig().String(config.EncryptDirKey)
	MasterKeyFileName = path.Join(encryptDir, MasterKeyFileName)
	PubKeyFileName = path.Join(encryptDir, PubKeyFileName)
	glog.Infof("MasterKeyFile: %v, PubKeyFile: %v", MasterKeyFileName, PubKeyFileName)
}

func encryptWithAttributeBased(text string, role string) (string, error) {
	var cmd *exec.Cmd
	var err error

	//write text into textFile
	encryptFileName := randStringRunes(12) + ".pdf"
	err = writeFile(encryptFileName, text)
	if err != nil {
		return "", err
	}
	defer os.Remove(encryptFileName)
	defer os.Remove(encryptFileName + ".cpabe")

	if !fileExists(PubKeyFileName) {
		return "", NoPubKey
	}

	//cpabe-enc pub_key security_report.pdf
	//encryptCmd := "echo -ne \"" + abe +"\" | cpabe-enc pub_key " + encryptFileName
	var attributePolicy string
	switch role {
	case OpServiceRole:
		attributePolicy = "(op_service = " + roleMap[role] + ")"
	case AdminRole:
		attributePolicy = "op_service or (admin = " + roleMap[role] + ")"
	case UserRole:
		attributePolicy = "op_service or admin or (user = " + roleMap[role] + ")"
	default:
		return "", NoRole
	}
	str := "cpabe-enc " + PubKeyFileName + " " + encryptFileName + " \"" + attributePolicy + "\""
	cmd = exec.Command("bash", "-c", str)
	if _, err = cmd.Output(); err != nil {
		return "", err
	}
	f, _ := os.OpenFile(encryptFileName+".cpabe", os.O_RDONLY, 0666)
	b, _ := ioutil.ReadAll(f)

	decoded := base64.StdEncoding.EncodeToString(b)
	return decoded, nil
}

func decryptWithAttributeBased(text string, role string) (string, error) {
	var cmd *exec.Cmd
	var err error
	var roleAttribute string
	var body []byte

	body, err = base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed")
	}
	text = string(body)

	filename := role + "_priv_key"
	if fileExists(filename) {
		err = os.Remove(filename)
		if err != nil {
			return "", err
		}
	}
	defer os.Remove(filename)
	switch role {
	case OpServiceRole:
		roleAttribute = "op_service 'op_service = " + roleMap[role] + "'"
	case AdminRole:
		roleAttribute = "admin 'admin = " + roleMap[role] + "'"
	case UserRole:
		roleAttribute = "'user = " + roleMap[role] + "'"
	default:
		return "", NoRole
	}

	if !fileExists(PubKeyFileName) {
		return "", NoPubKey
	}
	if !fileExists(MasterKeyFileName) {
		return "", NoMasterKey
	}
	//cpabe-keygen -o sara_priv_key pub_key master_key \
	genPrivacyKey := "cpabe-keygen -o " + filename + " " + PubKeyFileName + " " + MasterKeyFileName + " " + roleAttribute
	cmd = exec.Command("bash", "-c", genPrivacyKey)
	if _, err = cmd.Output(); err != nil {
		return "", err
	}

	encryptFile := randStringRunes(12) + ".pdf"
	decryptFile := encryptFile + ".cpabe"
	err = writeFile(decryptFile, text)
	if err != nil {
		return "", err
	}
	defer os.Remove(decryptFile)

	decryptCmd := strings.Join([]string{"cpabe-dec", PubKeyFileName, filename, decryptFile}, " ")
	cmd = exec.Command("bash", "-c", decryptCmd)
	if _, err = cmd.Output(); err != nil {
		//cannot decrypt, attributes in key do not satisfy policy
		glog.Infof("decrypt command failed: %s\n", err)
		return "", NoPermission
	}

	f, _ := os.OpenFile(encryptFile, os.O_RDONLY, 0666)
	b, _ := ioutil.ReadAll(f)
	os.Remove(encryptFile)

	return string(b), nil
}
