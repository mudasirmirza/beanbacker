package main

import (
	"beanstalk"
	"encoding/json"
	"encryption"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

// func to generate a random name based on datetime
func buildFileName() string {
	return time.Now().Format("20060102150405")
}

func encryptDataToFile(encryptionMethod encryption.IEncryptionMethod, filename string, data []byte) {
	dataFile, _ := os.Create(filename)
	defer dataFile.Close()
	encryptionOutput := encryptionMethod.EncryptData(data)
	dataFile.Write(encryptionOutput.EncryptedData)
	
	dataKeyFile, _ := os.Create(filename + "_dataKey")
	defer dataFile.Close()
	dataKeyFile.Write(encryptionOutput.EncryptedDataKey)
}

func decryptFile(encryptionMethod encryption.IEncryptionMethod, filename string) []byte {
	data, _ := ioutil.ReadFile(filename)
	return encryptionMethod.DecryptData(data)
}

func loadDataKey(filename string) []byte {
	data, _ := ioutil.ReadFile(filename)
	return data
}

func main() {

	destFile := flag.String("destFile", "", "Path to file where JSON data will be stored, default: \"\"")
	encryptionMethodName := flag.String("encryptionMethod", "AESCipherEncryption", "Choose between: "+strings.Join(encryption.GetAvailableEncryptionMethodNames(), ","))
	passPhrase := flag.String("passPhrase", "", "PassPhrase to encrypt file with, default: \"\"")
	kmsKeyArn := flag.String("kmsKeyArn", "", "Arn to our KMS key")
	decryptData := flag.Bool("decryptData", false, "Used to decrypt the data fetched from AWS BeanstalkEnvironments, default: false")
	encryptedDataKeyFile := flag.String("encryptedDataKeyFile", "", "If using AWS KMS encryption it is required")
	flag.Parse()
	var encryptionArgs = map[string]string{
		"passphrase": *passPhrase,
		"kmsKeyArn":  *kmsKeyArn,
	}
	if *encryptedDataKeyFile != "" {
		encryptionArgs["encryptedDataKey"] = string(loadDataKey(*encryptedDataKeyFile))
	}
	encryption.Init()
	encryptionMethod := encryption.AvailableEncryptionMethods[*encryptionMethodName]
	encryptionMethod.SetEncryptionArgs(encryptionArgs)

	// make sure destFile is defined in all cases
	if *destFile == "" && !*decryptData {
		fmt.Println("Destination file path not defined, data will not be saved in any file")
		fmt.Println("Please provide \"-destFile\" to write data to file.")
		fmt.Println("This utility does not print out data on STDOUT")
		os.Exit(1)
	}

	// passphrase is required in all cases
	if *passPhrase == "" {
		fmt.Println("Passphrase is required to encrypt the file containing all the environment variables")
		fmt.Println("Please provide \"-passPhrase\" to encrypt the file with.")
		os.Exit(1)
	}

	// destFile is required when decrypting the file
	if *destFile == "" && *decryptData {
		fmt.Println("Destination file path not defined, which file to decrypt")
		fmt.Println("Please provide \"-destFile\" to decrypts the data.")
		os.Exit(1)
	}

	// when decrypting, we don't need to perform any other operation
	if *decryptData {
		fileName := "./" + buildFileName() + ".json"
		dataFile, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Can not create temporary file ", fileName)
			log.Fatal(err)
		}
		dataFile.Write(decryptFile(encryptionMethod, *destFile))
		fmt.Println("Decrypted data written to ", fileName)
		os.Exit(0)
	}

	envDetails := beanstalk.FetchAllBeanstalkEnvVars()

	// creating a json out of the main slice
	b, _ := json.MarshalIndent(envDetails, "", "    ")

	// encrypting the json
	encryptDataToFile(encryptionMethod, *destFile, b)
}
