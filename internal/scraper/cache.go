package scraper

func (x *Scraper) CachedRecipe(sourceURL string) (recipe RecipeObject, ok bool) {
	// cache lookup to avoid HTTP transactions
	obj, ok := x.cache.Get(sourceURL)
	if !ok {
		return
	}
	recipe = obj.(RecipeObject)
	return
}

func (x *Scraper) addRecipeToCache(sourceURL string, recipe RecipeObject) (err error) {
	x.cache.Set(sourceURL, recipe, 1)
	x.cache.Wait()
	return
}
