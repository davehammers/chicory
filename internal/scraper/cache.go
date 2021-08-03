package scraper

func (x *Scraper) CachedRecipe(siteUrl string) (recipe RecipeObject, ok bool) {
	// cache lookup to avoid HTTP transactions
	obj, ok := x.cache.Get(siteUrl)
	if !ok {
		return
	}
	recipe = obj.(RecipeObject)
	return
}

func (x *Scraper) addRecipeToCache(siteUrl string, recipe RecipeObject) (err error) {
	x.cache.Set(siteUrl, recipe, 1)
	x.cache.Wait()
	return
}
