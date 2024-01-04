package crypt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestMd5(t *testing.T) {
	fmt.Println(Md5("123"))
}

func TestRas(t *testing.T) {
	SaveRsaKey(32)
}

func TestRsaEncoding(t *testing.T) {
	b, _ := RsaEncoding("test", "./publicKey.pem")
	fmt.Println(fmt.Sprintf("len %b", len(string(b))))
	c, _ := RsaDecoding(b, "./privateKey.pem")
	fmt.Println(fmt.Sprintf("!!!! %s", string(c)))
}

func TestAesEncoding(t *testing.T) {
	x := []byte("1234567890123456")
	encrypt, err := AesEncrypt([]byte("111"), x)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fmt.Sprintf("!!!! %s len %d", string(encrypt), len(string(encrypt))))
	decrypt, err := AesDecrypt(encrypt, x)
	fmt.Println(fmt.Sprintf("!!!! %s", string(decrypt)))
	if err != nil {
		return
	}
}

func TestAesEncoding1(t *testing.T) {
	orig := "hello world"
	key := "0123456789012345"
	fmt.Println("原文：", orig)
	encryptCode := AesEncryptBase64(orig, key)
	fmt.Println("密文：", encryptCode)
	decryptCode := AesDecryptBase64(encryptCode, key)
	fmt.Println("解密结果：", decryptCode)
}

func TestVkeySign(t *testing.T) {
	vkey := "b60d5103b32845048ebb3c0eb43f0118"
	signVkey := VkeyEncodingSign(vkey)
	fmt.Println(fmt.Sprintf("vkey sign %s", signVkey))
	vkeyn, err := VerifySign(signVkey)
	if err != nil {
		fmt.Println(fmt.Sprintf("vkey error %s", err))
		return
	}
	fmt.Println(fmt.Sprintf("vkey %s", vkeyn))
}

func TestCloudCreateSign(t *testing.T) {
	cloudSign := "sexxewqxx"
	sign := createSign([]string{"0335753700200000000001", "8662c0e6aeb1421b8208aa4fc6de9fb3"}, cloudSign)
	fmt.Println(sign)
}

func TestBinary(t *testing.T) {
	a := "hello, this is a first sentence."

	by := bytes.NewBuffer([]byte{})
	binary.Write(by, binary.BigEndian, int32(len(a)))

	var dataLen int32
	binary.Read(by, binary.BigEndian, &dataLen)
	fmt.Println(dataLen)
}

func TestBinaryStr(t *testing.T) {
	b := "hello, this is a second sentence."
	by := bytes.NewBuffer([]byte{})
	binary.Write(by, binary.BigEndian, []byte(b))

	var dataStr = make([]byte, len(b))
	binary.Read(by, binary.BigEndian, &dataStr)
	fmt.Println(string(dataStr))
}

func TestStruct(t *testing.T) {
	type User struct {
		Name [30]byte
		Age  uint32
	}
	user := User{}
	copy(user.Name[:], "John Wilson")
	user.Age = 30
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, &user)

	var user1 User
	binary.Read(buffer, binary.BigEndian, &user1)
	fmt.Println(user1)
}
