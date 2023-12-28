# abandonship

Simple Go app for notifying via Pushover. Intended to be used in a shutdown script from a GCP instance. Looks up credentials from a Secret Manager secret.

## Usage

```
abandonship -s <secret-name> -m '<message>'
```

## Assumptions

* The runtime environment has default credentials set with `cloud-platform` scope
* The secret content is a YAML document with `token` and `user` keys set
* Latest version of the secret is fetched

## Why

Because looking up secrets to fetch Pushover credentials and using the credentials to send a message on shutdown doesn't seem to be reliable when done with Bash.
