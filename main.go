package main

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"log"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/alexflint/go-arg"
	"github.com/gregdel/pushover"
	"gopkg.in/yaml.v3"
)

const (
	appVersion     = "0.1.0"
	name           = "abandonship"
	scope          = "cloud-platform"
	defaultVersion = "latest"
	secretSpec     = "projects/%s/secrets/%s/versions/%s"
)

type args struct {
	Message       string `arg:"required,-m,--message" help:"Message body"`
	Secret        string `arg:"required,-s,--secret" help:"Secret name"`
	SecretVersion string `arg:"-v,--secret-version" help:"Secret version"`
}

func (args) Version() string {
	return fmt.Sprintf("%s v%s", name, appVersion)
}

type credentials struct {
	Token string `yaml:"token"`
	User  string `yaml:"user"`
}

func readSecret(ctx context.Context, secret, version string) (out []byte, err error) {
	creds, err := google.FindDefaultCredentials(ctx, scope)
	if err != nil {
		return out, fmt.Errorf("unable to construct credentials: %v", err)
	}

	project := creds.ProjectID
	if project == "" {
		return out, fmt.Errorf("unable to determine project from default credentials")
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return out, fmt.Errorf("error creating secret manager client: %v", err)
	}
	defer client.Close()

	if version == "" {
		version = defaultVersion
	}
	secretName := fmt.Sprintf(secretSpec, project, secret, version)
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return out, fmt.Errorf("error accessing secret %s: %v", secretName, err)
	}

	return result.Payload.Data, nil
}

func main() {
	var parsed args
	arg.MustParse(&parsed)

	secret, err := readSecret(context.Background(), parsed.Secret, parsed.SecretVersion)
	if err != nil {
		log.Fatal(err)
	}

	var creds credentials
	err = yaml.Unmarshal(secret, &creds)
	if err != nil {
		log.Fatal(err)
	}

	app := pushover.New(creds.Token)
	recipient := pushover.NewRecipient(creds.User)
	message := pushover.NewMessage(parsed.Message)

	_, err = app.SendMessage(message, recipient)
	if err != nil {
		log.Fatal(err)
	}
}
