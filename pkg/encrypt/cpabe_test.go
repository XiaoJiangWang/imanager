package encrypt

import (
	"os"
	"os/exec"
	"testing"
)

func TestUserEncryptAndDecrypt(t *testing.T) {
	//cpabe-setup
	cmd := exec.Command("cpabe-setup")
	if _, err := cmd.Output(); err != nil {
		t.Logf("gen pub key failed, err: %v", err)
		t.Fail()
		return
	}
	defer func() {
		os.Remove(PubKeyFileName)
		os.Remove(MasterKeyFileName)
	}()

	text := "Hello World!"
	encryptData, err := encryptWithAttributeBased(text, UserRole)
	if err != nil {
		t.Logf("encrypt failed, err: %v", err)
		t.Fail()
		return
	}
	data, err := decryptWithAttributeBased(encryptData, UserRole)
	if err != nil {
		t.Logf("decrypt failed, err: %v", err)
		t.Fail()
		return
	}
	if text != data {
		t.Logf("Encryption and decryption are not equivalent, text: %v, data: %v", text, data)
		t.Fail()
		return
	}
}

func TestAdminEncryptAndDecrypt(t *testing.T) {
	//cpabe-setup
	cmd := exec.Command("cpabe-setup")
	if _, err := cmd.Output(); err != nil {
		t.Logf("gen pub key failed, err: %v", err)
		t.Fail()
		return
	}
	defer func() {
		os.Remove(PubKeyFileName)
		os.Remove(MasterKeyFileName)
	}()

	text := "Hello World!"
	encryptData, err := encryptWithAttributeBased(text, AdminRole)
	if err != nil {
		t.Logf("encrypt failed, err: %v", err)
		t.Fail()
		return
	}
	data, err := decryptWithAttributeBased(encryptData, AdminRole)
	if err != nil {
		t.Logf("decrypt failed, err: %v", err)
		t.Fail()
		return
	}
	if text != data {
		t.Logf("Encryption and decryption are not equivalent, text: %v, data: %v", text, data)
		t.Fail()
		return
	}
}

func TestOpServiceEncryptAndDecrypt(t *testing.T) {
	//cpabe-setup
	cmd := exec.Command("cpabe-setup")
	if _, err := cmd.Output(); err != nil {
		t.Logf("gen pub key failed, err: %v", err)
		t.Fail()
		return
	}
	defer func() {
		os.Remove(PubKeyFileName)
		os.Remove(MasterKeyFileName)
	}()

	text := "Hello World!"
	encryptData, err := encryptWithAttributeBased(text, OpServiceRole)
	if err != nil {
		t.Logf("encrypt failed, err: %v", err)
		t.Fail()
		return
	}
	data, err := decryptWithAttributeBased(encryptData, OpServiceRole)
	if err != nil {
		t.Logf("decrypt failed, err: %v", err)
		t.Fail()
		return
	}
	if text != data {
		t.Logf("Encryption and decryption are not equivalent, text: %v, data: %v", text, data)
		t.Fail()
		return
	}
}

func TestExceedAuthEncryptAndDecrypt(t *testing.T) {
	//cpabe-setup
	cmd := exec.Command("cpabe-setup")
	if _, err := cmd.Output(); err != nil {
		t.Logf("gen pub key failed, err: %v", err)
		t.Fail()
		return
	}
	defer func() {
		os.Remove(PubKeyFileName)
		os.Remove(MasterKeyFileName)
	}()

	text := "Hello World!"
	encryptData, err := encryptWithAttributeBased(text, UserRole)
	if err != nil {
		t.Logf("encrypt failed, err: %v", err)
		t.Fail()
		return
	}

	// use admin role
	data, err := decryptWithAttributeBased(encryptData, AdminRole)
	if err != nil {
		t.Logf("decrypt failed, err: %v", err)
		t.Fail()
		return
	}
	if text != data {
		t.Logf("Encryption and decryption are not equivalent, text: %v, data: %v", text, data)
		t.Fail()
		return
	}

	// use op_service role
	data, err = decryptWithAttributeBased(encryptData, OpServiceRole)
	if err != nil {
		t.Logf("decrypt failed, err: %v", err)
		t.Fail()
		return
	}
	if text != data {
		t.Logf("Encryption and decryption are not equivalent, text: %v, data: %v", text, data)
		t.Fail()
		return
	}

	// use op service encrypt
	encryptData, err = encryptWithAttributeBased(text, OpServiceRole)
	if err != nil {
		t.Logf("encrypt failed, err: %v", err)
		t.Fail()
		return
	}
	data, err = decryptWithAttributeBased(encryptData, AdminRole)
	if err != NoPermission {
		t.Logf("admin shouldn't decrypt op service data, err: %v", err)
		t.Fail()
		return
	}
	data, err = decryptWithAttributeBased(encryptData, UserRole)
	if err != NoPermission {
		t.Logf("user shouldn't decrypt op service data, err: %v", err)
		t.Fail()
		return
	}
}