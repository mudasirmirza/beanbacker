package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// struct for holding key value pair of environment variables
type EnvVarDetails   struct {
	VariableName string `json:"VariableName"`
	VariableValue string `json:"VariableValue"`
}

// struct for holding environment detail of individual environments
type EnvVar struct {
	EnvironmentName string `json:"EnvironmentName"`
	EvnVars		[]EnvVarDetails	`json:"EnvVarDetails"`
}

// struct for holding environment details of all environments of all applications
type AllAppsEnvVars struct {
	EnvironmentDetails	[]EnvVar	`json:"EnvironmentDetails"`
}

// func to perform error checking using AWS error
func checkErr(err error) {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
}

// func to check AWS env vars
func checkAWSEnvVar() bool{
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("AWS_REGION")
	if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" {
		return false
	}
	return true
}

// Func to get config options for the given env and app
func getEnvConfigSettings(envName *string, appName *string, sess *session.Session) ([]*elasticbeanstalk.ConfigurationSettingsDescription) {
	svc := elasticbeanstalk.New(sess)
	input := &elasticbeanstalk.DescribeConfigurationSettingsInput{
		ApplicationName: appName,
		EnvironmentName: envName,
	}
	res, err := svc.DescribeConfigurationSettings(input)
	checkErr(err)
	return res.ConfigurationSettings
}

// func to create a random hash
// https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Func to encrypt text
// https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

// Func to decrypt text
// https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

// func to encrypt data to file
// https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func encryptFile(filename string, data []byte, passphrase string) {
	f, _ := os.Create(filename)
	defer f.Close()
	f.Write(encrypt(data, passphrase))
}

// func decrypt data from file
// https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func decryptFile(filename string, passphrase string) []byte {
	data, _ := ioutil.ReadFile(filename)
	return decrypt(data, passphrase)
}

// func to generate a random name based on datetime
func buildFileName() string {
	return time.Now().Format("20060102150405")
}

func main() {

	destFile := flag.String("destFile", "", "Path to file where JSON data will be stored, default: \"\"")
	passPhrase := flag.String("passPhrase", "", "PassPhrase to encrypt file with, default: \"\"")
	decryptData := flag.Bool("decryptData", false, "Used to decrypt the data fetched from AWS BeanstalkEnvironments, default: false")
	flag.Parse()

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

	// check if AWS credentials are present as environment variables
	if !checkAWSEnvVar() {
		log.Fatal("AWS environment variables not defiled")
		os.Exit(1)
	}

	// when decrypting, we don't need to perform any other operation
	if *decryptData {
		fileName := "./"+buildFileName()+".json"
		dataFile, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Can not create temporary file ", fileName)
			log.Fatal(err)
		}
		dataFile.Write(decryptFile(*destFile, *passPhrase))
		fmt.Println("Decrypted data written to ", fileName)
		os.Exit(0)
	}

	// creating a new session which will be used in all AWS API calls
	sess, err := session.NewSession()
	checkErr(err)
	// creating a elasticbeanstalk service object
	svc := elasticbeanstalk.New(sess)
	// input parameters required when calling describe environments
	envInput := &elasticbeanstalk.DescribeEnvironmentsInput{}
	// describing beanstalk environments
	res, err := svc.DescribeEnvironments(envInput)
	checkErr(err)

	// initializing variables
	// variable to hold environment name
	var envName *string
	// variable to hold environment variables of all environments in all applications
	var envDetails = AllAppsEnvVars{}
	// variable to hold environment variable of individual environments
	var envVariables = []EnvVar{}

	// looping over all environments
	for _, v := range res.Environments {
		// setting environment name variable
		envName = v.EnvironmentName
		// getting configuration settings for the individual environment
		confOpt := getEnvConfigSettings(v.EnvironmentName, v.ApplicationName, sess)
		// instantiating and initializing variable to hold environment variables of individual environment
		envVar := EnvVar{EnvironmentName: *envName}
		// initializing slice to hold all environment variables of an individual environment
		var varDetails = []EnvVarDetails{}
		// looping over all the configuration option settings
		for _, v := range confOpt[0].OptionSettings {
			// making sure we only perform action on application environment variables
			if *v.Namespace == "aws:elasticbeanstalk:application:environment" {
				// here we append all the environment variables in a slice
				varDetails = append(varDetails, EnvVarDetails{VariableName: *v.OptionName, VariableValue: *v.Value})
			}
		}
		// adding the complete slice to the main slice which holds details for an individual environment
		envVar.EvnVars = varDetails
		// merging slices holding details of individual environments in single slice
		envVariables = append(envVariables, envVar)
	}
	// creating a top level slice to hold details of all environments in individual slices
	envDetails.EnvironmentDetails = envVariables
	// creating a json out of the main slice
	b, _ := json.MarshalIndent(envDetails, "", "    ")
	// encrypting the json
	encryptFile(*destFile, b, *passPhrase)
}

