package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// 保存生成的公钥和密钥
func SaveRsaKey(bits int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		fmt.Println(err)
		return err
	}
	publicKey := privateKey.PublicKey

	// 使用x509标准对私钥进行编码，AsN.1编码字符串
	x509Privete := x509.MarshalPKCS1PrivateKey(privateKey)
	// 使用x509标准对公钥进行编码，AsN.1编码字符串
	x509Public := x509.MarshalPKCS1PublicKey(&publicKey)

	// 对私钥封装block 结构数据
	blockPrivate := pem.Block{Type: "private key", Bytes: x509Privete}
	// 对公钥封装block 结构数据
	blockPublic := pem.Block{Type: "public key", Bytes: x509Public}

	// 创建存放私钥的文件
	privateFile, errPri := os.Create("privateKey.pem")
	if errPri != nil {
		return errPri
	}
	defer privateFile.Close()
	pem.Encode(privateFile, &blockPrivate)

	// 创建存放公钥的文件
	publicFile, errPub := os.Create("publicKey.pem")
	if errPub != nil {
		return errPub
	}
	defer publicFile.Close()
	pem.Encode(publicFile, &blockPublic)

	return nil

}

// 加密
func RsaEncoding(src, filePath string) ([]byte, error) {

	srcByte := []byte(src)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return srcByte, err
	}

	// 获取文件信息
	fileInfo, errInfo := file.Stat()
	if errInfo != nil {
		return srcByte, errInfo
	}
	// 读取文件内容
	keyBytes := make([]byte, fileInfo.Size())
	// 读取内容到容器里面
	file.Read(keyBytes)

	// pem解码
	block, _ := pem.Decode(keyBytes)

	// x509解码
	publicKey, errPb := x509.ParsePKCS1PublicKey(block.Bytes)
	if errPb != nil {
		return srcByte, errPb
	}

	// 使用公钥对明文进行加密

	retByte, errRet := rsa.EncryptPKCS1v15(rand.Reader, publicKey, srcByte)
	if errRet != nil {
		return srcByte, errRet
	}

	return retByte, nil

}

// 解密
func RsaDecoding(srcByte []byte, filePath string) ([]byte, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return srcByte, err
	}
	// 获取文件信息
	fileInfo, errInfo := file.Stat()
	if errInfo != nil {
		return srcByte, errInfo
	}
	// 读取文件内容
	keyBytes := make([]byte, fileInfo.Size())
	// 读取内容到容器里面
	_, _ = file.Read(keyBytes)
	// pem解码
	block, _ := pem.Decode(keyBytes)
	// x509解码
	privateKey, errPb := x509.ParsePKCS1PrivateKey(block.Bytes)
	if errPb != nil {
		return keyBytes, errPb
	}
	// 进行解密
	retByte, errRet := rsa.DecryptPKCS1v15(rand.Reader, privateKey, srcByte)
	if errRet != nil {
		return srcByte, errRet
	}
	return retByte, nil
}

func main() {
	//err := SaveRsaKey(2048)
	//if err != nil {
	//	fmt.Println("KeyErr",err)
	//}
	msg, err := RsaEncoding("FanOne", "publicKey.pem")
	fmt.Println("msg", msg)
	if err != nil {
		fmt.Println("err1", err)
	}
	msg2, err := RsaDecoding(msg, "privateKey.pem")
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("msg2", string(msg2))
}
