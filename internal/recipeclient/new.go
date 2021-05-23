// This package is the HTTP client that will connect with a website and extract the recipe information

package recipeclient

// RecipeClientIn - input parameters for a RecipeClient
type RecipeClientIn struct {
	Client RClient
}

// RecipeClient - internal structure to this package. Typically no fields are exported.
type RecipeClient struct {
	client RClient
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

// NewClient - allocate a *RecipeClient
func (x *RecipeClientIn) NewClient() *RecipeClient {
	return &RecipeClient{
		client: x.Client,
	}
}
