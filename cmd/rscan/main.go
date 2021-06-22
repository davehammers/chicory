// entry point for the recipe scanner

package main

import (
	"rscan/internal/config"
)

func main() {
	config.New().
		ConfigRedis().
		ConfigRecipeClient().
		ConfigServer().
		Start()
}
