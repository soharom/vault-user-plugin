package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/sdk/plugin"

	"github.com/hashicorp/vault-secret-user/version"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()

	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	if err := plugin.ServeMultiplex(&plugin.ServeOpts{
		BackendFactoryFunc: Factory,
		// set the TLSProviderFunc so that the plugin maintains backwards
		// compatibility with Vault versions that don’t support plugin AutoMTLS
		TLSProviderFunc: tlsProviderFunc,
	}); err != nil {
		log.Fatal(err)
	}
}

func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(c)
	if err := b.Setup(ctx, c); err != nil {
		return nil, err
	}
	return b, nil
}

type backend struct {
	*framework.Backend
}

func Backend(c *logical.BackendConfig) *backend {
	var b backend

	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{".*"},
		},
		Paths: []*framework.Path{
			{
				Pattern: ".*",
				Fields: map[string]*framework.FieldSchema{
					"password": {
						Type:        framework.TypeString,
						Description: "This password filed",
					},
					"username": {
						Type:        framework.TypeString,
						Description: "This is username field",
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.UpdateOperation: b.handleWrite,
					logical.ReadOperation:   b.handleRead,
					logical.ListOperation:   b.handleList,
				},
			},
		},
		RunningVersion: "v" + version.Version,
	}

	return &b
}

func (b *backend) handleWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	key := req.Path

	value := data.Get("password").(string)
	username := data.Get("username").(string)

	storedData := map[string]string{
		"username": username,
		"password": value,
	}

	jsonData, err := json.Marshal(storedData)
	if err != nil {
		b.Logger().Error("Failed to serialize the data", "error", err)
		return nil, err
	}

	entry := &logical.StorageEntry{
		Key:   key,
		Value: jsonData,
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		b.Logger().Error("Failed to store the secret", "error", err)
		return nil, err
	}

	b.Logger().Info("Data has been stored successfully", "key", key, "data", storedData)

	return &logical.Response{}, nil

}

func (b *backend) handleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	key := req.Path

	entry, err := req.Storage.Get(ctx, key)
	if err != nil {
		b.Logger().Error("Could not list the value")
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var storedData map[string]interface{}
	if err := json.Unmarshal(entry.Value, &storedData); err != nil {
		b.Logger().Error("Failed to deserialize the data", "error", err)
		return nil, err
	}

	b.Logger().Info("The value is listed ")
	return &logical.Response{
		Data: storedData,
	}, nil
}

func (b *backend) handleList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, req.Path)
	if err != nil {
		b.Logger().Error("Error listing the data", "error", err)
		return nil, err
	}

	if len(entries) == 0 {
		b.Logger().Info("No entries found", "path", req.Path)
		return &logical.Response{}, nil
	}

	b.Logger().Info("Entries listed successfully", "path", req.Path, "entries", entries)
	return logical.ListResponse(entries), nil
}
