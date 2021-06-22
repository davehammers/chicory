// This package is the HTTP client that will connect with a website and extract the recipe information

package recipeclient

import (
	"context"
	"github.com/go-redis/redis/v8"
)

// RecipeClientIn - input parameters for a RecipeClient
type RecipeClientIn struct {
	Client RClient
	Redis  *redis.Client
}

// RecipeClient - internal structure to this package. Typically no fields are exported.
type RecipeClient struct {
	client RClient
	redis  *redis.Client
	ctx    context.Context
}

// RecipeClientIn - allocate a new RecipeClientIn. Returns *RecipeClientIn
func New() *RecipeClientIn {
	return &RecipeClientIn{}
}

// SetClient - used to set the HTTP client for this package
// usage:
//	client := recipeclient.New().
//		SetClient(myClient).
//		NewClient()
func (x *RecipeClientIn) SetClient(in RClient) *RecipeClientIn {
	x.Client = in
	return x
}
func (x *RecipeClientIn) SetRedis(in *redis.Client) *RecipeClientIn {
	x.Redis = in
	return x
}

// NewClient - allocate a *RecipeClient
func (x *RecipeClientIn) NewClient() *RecipeClient {
	return &RecipeClient{
		client: x.Client,
		redis:  x.Redis,
		ctx:    context.Background(),
	}
}
