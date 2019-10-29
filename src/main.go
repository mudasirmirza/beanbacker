package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"log"
	"time"
	"encryption"
	"encoding/json"
	"os"
	"io/ioutil"
	"strings"
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

// func to generate a random name based on datetime
func buildFileName() string {
	return time.Now().Format("20060102150405")
}

func encryptDataToFile(encryptionMethod encryption.IEncryptionMethod, filename string, data []byte) {
	f, _ := os.Create(filename)
	defer f.Close()
	f.Write(encryptionMethod.EncryptData(data))
}

func decryptFile(encryptionMethod encryption.IEncryptionMethod, filename string) []byte {
	data, _ := ioutil.ReadFile(filename)
	return encryptionMethod.DecryptData(data)
}

func main() {

	destFile := flag.String("destFile", "", "Path to file where JSON data will be stored, default: \"\"")
	encryptionMethodName := flag.String("encryptionMethod", "AESCipherEncryption", "Choose between: " + strings.Join(encryption.GetAvailableEncryptionMethodNames(), ","))
	passPhrase := flag.String("passPhrase", "", "PassPhrase to encrypt file with, default: \"\"")
	kmsKeyArn := flag.String("kmsKeyArn", "", "Arn to our KMS key")
	decryptData := flag.Bool("decryptData", false, "Used to decrypt the data fetched from AWS BeanstalkEnvironments, default: false")

	flag.Parse()
	var encryptionArgs = map[string]string{
		"passphrase": *passPhrase,
		"kmsKeyArn": *kmsKeyArn,
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
		fileName := "./"+buildFileName()+".json"
		dataFile, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Can not create temporary file ", fileName)
			log.Fatal(err)
		}
		dataFile.Write(decryptFile(encryptionMethod, *destFile))
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
	encryptDataToFile(encryptionMethod, *destFile, b)
}
