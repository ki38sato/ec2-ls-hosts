package creds

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"gopkg.in/ini.v1"
	"os"
	"strings"
)

const (
	SHARED_CREDS_FILENAME  = "/.aws/credentials"
	SHARED_CREDS_PROFILE   = "default"
	SHARED_CONFIG_FILENAME = "/.aws/config"
)

// creds "env" or "shared" or "ec2"
func SelectCredentials(creds string, profiles string) ([]*credentials.Credentials, error) {
	sharedCredsFile := os.Getenv("HOME") + SHARED_CREDS_FILENAME
	switch creds {
	case "env":
		return []*credentials.Credentials{
			credentials.NewEnvCredentials(),
		}, nil
	case "shared":
		return []*credentials.Credentials{
			credentials.NewSharedCredentials(sharedCredsFile, SHARED_CREDS_PROFILE),
		}, nil
	case "shared_multi":
		credentials_list := make([]*credentials.Credentials, 0)
		for _, profile := range strings.Split(profiles, ",") {
			credentials_list = append(credentials_list, credentials.NewSharedCredentials(sharedCredsFile, profile))
		}
		return credentials_list, nil
	case "ec2":
		return []*credentials.Credentials{
			credentials.NewCredentials(&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(session.New()),
			}),
		}, nil
	case "":

		providers := []credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Filename: sharedCredsFile,
				Profile:  SHARED_CREDS_PROFILE,
			},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(session.New()),
			}}
		return []*credentials.Credentials{credentials.NewChainCredentials(providers)}, nil
	default:
		return nil, errors.New("Unknown creds name: " + creds)
	}
}

func LoadAwsConfig() (*ini.File, error) {
	return ini.Load(os.Getenv("HOME") + SHARED_CONFIG_FILENAME)
}
