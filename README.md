# Welcome to Dotenc

Dotenc is a Go application to encrypt your dotenv files
so that they can be securely shared in git repositories.

Inspired by `ejson` ruby gem https://github.com/Shopify/ejson.

The main difference between ejson and dotenc is that
dotenc works on dotenv files while ejson works on json files.

## Help

If you run dotenc without any arguments it will print help message with the usage instructions.

    Dotenc is a small library to manage encrypted secrets using asymetric encryption.

    Usage:
      dotenc [command]

    Available Commands:
      encrypt [env]          Encrypt given environment file .env.[env]
      e [env]                Shortcut for encrypt
      decrypt [env]          Decrypt given environment file .env.[env] and print to STDOUT
      d [env]                Shortcut for decrypt
      generate               Generate new public and private key
      g                      Shortcut for generate
      exec [env] [command]   Decrypt and load env variables from .env.[env] file and run program [command]

## Usage

### Encryption key

Generate new public and private key:

    dotenc generate

Example of generated output:

    Public key: 851d1baf941bfe98a54c87790a74ea1a87b068c8d535ca7969f229cd996e2d7c
    Private key: e2e4274ca2dc5386213adc2fc62d87a2a5c38afa4ab949a49726d7ebcf0c75dc

    Add this line on top of your dotfile:
    # public_key: 851d1baf941bfe98a54c87790a74ea1a87b068c8d535ca7969f229cd996e2d7c

    Add this line to your .dotenc file:
    851d1baf941bfe98a54c87790a74ea1a87b068c8d535ca7969f229cd996e2d7c: e2e4274ca2dc5386213adc2fc62d87a2a5c38afa4ab949a49726d7ebcf0c75dc

    Remember to ignore .dotenc in your version control system! You can use following command:
    echo ".dotenc" >> .gitignore

Add generated public key to top of your env file. Example:

    # public_key: 851d1baf941bfe98a54c87790a74ea1a87b068c8d535ca7969f229cd996e2d7c

Add private key to `.dotenc` file. Example:

    851d1baf941bfe98a54c87790a74ea1a87b068c8d535ca7969f229cd996e2d7c: e2e4274ca2dc5386213adc2fc62d87a2a5c38afa4ab949a49726d7ebcf0c75dc

Add secrets to your env file:

    echo "FOO=bar" >> .env.production
    echo "export ENV=production" >> .env.production

### Git

Add your `.dotenc` encryption key to `.gitignore` so that it is not commited to repository.

    echo ".dotenc" >> .gitignore

### Encrypt env file

Encrypt your secret env file:

    dotenc encrypt production
    dotenc e production
    cat .env.production

Example content of generated encrypted env file `.env.production`:

    # public_key: 851d1baf941bfe98a54c87790a74ea1a87b068c8d535ca7969f229cd996e2d7c

    FOO=EJ[1:z4M3hY5e+xyfuxVCqG2rGvawmwBimvkJRpi5JYyLD0o=:I7P2CGyBPkS3dP7Sh/3VYFg2Aa0T6VdX:oqEhBaNMA54bDhOotPqVsqBH1g==]
    export ENV=EJ[1:z4M3hY5e+xyfuxVCqG2rGvawmwBimvkJRpi5JYyLD0o=:fPfzBgXMlFo48KxIS4wpAembxuVUgPjA:L+3ZdxinpRixIn5IsTtDkc6AwaFu6SoVX14=]

### Decrypt env file

Now that the files are encrypted, you can decrypt them to STDOUT:

    dotenc decrypt production
    dotenc d production

Example content of decrypted `.env.production` file:

    # public_key: 851d1baf941bfe98a54c87790a74ea1a87b068c8d535ca7969f229cd996e2d7c

    FOO=bar
    export ENV=production

### Executing Commands

Dotenc also provides a way to decrypt and load env files to execute any command.

    dotenv exec production mycommand with args

## Encryption

Dotenc uses ejson crypto https://github.com/Shopify/ejson/blob/master/crypto/crypto.go to encrypt env values.

Encryption key is read from a `.dotenc` file from the current directory.
That file should never be commited to your repository.

## Requirements

- Developed with Go version go1.16.3 darwin/amd64

## Installation instructions

This program can be installed easily if you have the go language installed on your system.

    go get -u github.com/atrzaska/dotenc

Make sure that you have your go bin folder in your path. Add following line to your shell RC file.

    export PATH="~/go/bin:$PATH"

## Building locally

    go build

## Exports

To provide copy paste support from shell scripts, export keywords will be ignored, when reading dotenv files.

With that said, both versions of following environment variable definition will work just fine:

Dotenv syntax

    NODE_ENV=development

Shell export syntax

    export NODE_ENV=development

## Licence

MIT
