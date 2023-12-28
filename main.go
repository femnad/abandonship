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
	name    = "abandonship"
	version = "0.1.0"
)

type args struct {
	Message string `arg:"required,-m,--message" help:"Message body"`
	Secret  string `arg:"required,-s,--secret" help:"Secret name"`
}

func (args) Version() string {
	return fmt.Sprintf("%s v%s", name, version)
}

type credentials struct {
	Token string `yaml:"token"`
	User  string `yaml:"user"`
}

func readSecret(ctx context.Context, secret string) (out []byte, err error) {
	creds, err := google.FindDefaultCredentials(ctx, "cloud-platform")
	if err != nil {
		return out, fmt.Errorf("unable to construct credentials: %v", err)
	}

	project := creds.ProjectID
	if project == "" {
		return out, fmt.Errorf("unable to determine project from default credentials")
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("error creating secret manager client: %v", err)
	}
	defer client.Close()

	secretName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", project, secret)
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return []byte{}, fmt.Errorf("error accessing secret %s: %v", secretName, err)
	}

	return result.Payload.Data, nil
}

func main() {
	var parsed args
	arg.MustParse(&parsed)

	secret, err := readSecret(context.Background(), parsed.Secret)
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
