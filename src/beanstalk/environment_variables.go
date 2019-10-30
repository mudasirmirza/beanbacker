package beanstalk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"utils"
)

// struct for holding key value pair of environment variables
type EnvVarDetails struct {
	VariableName  string `json:"VariableName"`
	VariableValue string `json:"VariableValue"`
}

// struct for holding environment detail of individual environments
type EnvVar struct {
	EnvironmentName string          `json:"EnvironmentName"`
	EvnVars         []EnvVarDetails `json:"EnvVarDetails"`
}

// struct for holding environment details of all environments of all applications
type AllAppsEnvVars struct {
	EnvironmentDetails []EnvVar `json:"EnvironmentDetails"`
}

// Func to get config options for the given env and app
func getEnvConfigSettings(envName *string, appName *string, sess *session.Session) []*elasticbeanstalk.ConfigurationSettingsDescription {
	svc := elasticbeanstalk.New(sess, aws.NewConfig())
	input := &elasticbeanstalk.DescribeConfigurationSettingsInput{
		ApplicationName: appName,
		EnvironmentName: envName,
	}
	res, err := svc.DescribeConfigurationSettings(input)
	utils.CheckErr(err)
	return res.ConfigurationSettings
}

func FetchAllBeanstalkEnvVars(existingSession *session.Session) AllAppsEnvVars {
	sess, err := session.NewSession()
	utils.CheckErr(err)
	if existingSession != nil {
		sess = existingSession
	}
	svc := elasticbeanstalk.New(sess)
	// input parameters required when calling describe environments
	envInput := &elasticbeanstalk.DescribeEnvironmentsInput{}
	// describing beanstalk environments
	res, err := svc.DescribeEnvironments(envInput)
	utils.CheckErr(err)

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
	return envDetails
}
