# Welcome to Dotenc

Dotenc is a Go application to encrypt your dotenv files so that they can be securely shared in git repositories.

## Examples:

Add your secret env file:

    echo "FOO=bar" >> .env.production
    echo "export ENV=production" >> .env.production

Initialize your encryption key with:

    pwgen -N 1 -s 100 > .dotenc

Encrypt your secret env file:

    dotenc encrypt production
    cat .env.production.enc

Example content of generated encrypted env file `.env.production.enc`:

    FOO=c82426d23fbc40dfdce1a0c53a888b161f4b1807122ed4938ab0650a525489
    export ENV=c67fd814ff05ffb546dba21ec787465f092cd9e5f8a384ec2de6de00e19a497372ddbc717b9e

## Encryption

Dotenc uses AES to encrypt env values and MD5 for hashing the password.

## Requirements

- Developed with `go version go1.16.3 darwin/amd64`

## Installation instructions
