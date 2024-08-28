

set -ex
PLUGIN_NAME=  "vault-secret-user"
GOOS=linux go build -o $PLUGIN_NAME cmd/vault-secret-user/main.go

docker kill vaultplg 2>/dev/null || true
tmpdir=$(mktemp -d vaultplgXXXXXX)
mkdir "$tmpdir/data"
docker pull hashicorp/vault
docker run --rm -d -p8200:8200 --name vaultplg -v C:\\Users\\admin\\Desktop\\projects\\vault-plugin-auth:/example --cap-add=IPC_LOCK -e 'VAULT_LOCAL_CONFIG=
{
  "backend": {"file": {"path": "/data"}},
  "listener": [{"tcp": {"address": "0.0.0.0:8200", "tls_disable": true}}],
  "plugin_directory": "/example",
  "log_level": "debug",
  "ui": "true",
  "disable_mlock": true,
  "api_addr": "http://localhost:8200"
}
' hashicorp/vault server

sleep 1
mkdir /data && \
chown vault:vault /data && \
export VAULT_ADDR=http://localhost:8200 && \

initoutput=$(vault operator init -key-shares=1 -key-threshold=1 -format=json) && \
apk add jq && \
vault operator unseal $(echo "$initoutput" | jq -r .unseal_keys_hex[0]) && \

export VAULT_TOKEN=$(echo "$initoutput" | jq -r .root_token) && cd /example && \

vault write sys/plugins/catalog/secret/example-auth-plugin \
    sha_256=$(sha256sum $PLUGIN_NAME| cut -d' ' -f1) \
    command="$PLUGIN_NAME" && \

vault secrets enable \
    -path="example" \
    -plugin-name="example-auth-plugin" \
    -plugin-version=0.2.0 \
    plugin

