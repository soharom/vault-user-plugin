# User KV Vault Plugin

This plugin utilizes the Vault secrets engine to store arbitrary key-value data, such as usernames and passwords for a given user. It serves as an example to demonstrate how to create and use a custom secrets plugin in Vault.

## Usage

To test and run this plugin, you can use the provided `docker.sh` script. This script sets up a Docker environment for running the Vault server with the plugin enabled.

### Running the Plugin

First, make sure you have Docker installed on your machine. Then, you can run the following command:

```bash
./docker.sh
