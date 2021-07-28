# Chicory - recipe scanner (scraper)

This is a simple application that scans URLs for recipes and displays the ingredients in JSON format.

## Github

The application is located in GitHub at https://github.com/davehammers/chicory.

The application follows the Golang standard project layout, described here https://github.com/golang-standards/project-layout

```
.
├── bin
├── cmd
│   └── scraper
└── internal
    ├── config
    ├── httpserver
    └── recipeclient
        └── testdata
```

## Building scraper

At the top level directory enter:

```sh
make install
```

The application will compile, run unit tests and install the binary in:

```sh
bin/scraper
```

## Running scraper

Rscan creates an HTTP web server on port 8080. To start the application: from a shell, enter

`bin/scraper`

Scan supports 2 endpoint URLs:

- `/scrape`
- `/scrapeall`

### /scrape

The /scrape endpoint accepts a query parameter `url` and extracts any recipe ingredients, then displays those to the browser.

E.g. http://0.0.0.0:8080/scrape?url=https://www.bettycrocker.com/recipes/slow-cooker-family-favorite-chili/c6f4c4e2-8298-4d9d-8f21-d4ca19d35cb7

```json
{
    "ingredients": [
        "2  lb lean (at least 80%) ground beef",
        "1  large onion, chopped (1 cup)",
        "2  cloves garlic, finely chopped",
        "1  can (28 oz) Muir Glen™ organic diced tomatoes",
        "1  can (16 oz) chili beans in sauce, undrained",
        "1  can (15 oz) Muir Glen™ organic tomato sauce",
        "2  tablespoons chili powder",
        "1 1/2  teaspoons ground cumin",
        "1/2  teaspoon salt",
        "1/2  teaspoon pepper"
    ]
}
```

### /scrapeall

The /scrapeall endpoint accepts a query  parameter `url` and extracts the schema.org information for the recipe, then displays the information to the browser.

E.g. http://0.0.0.0:8080/scrapeall?url=https://www.bettycrocker.com/recipes/slow-cooker-family-favorite-chili/c6f4c4e2-8298-4d9d-8f21-d4ca19d35cb7

```json
[
    {
        "@context": "http://schema.org/",
        "@type": "Recipe",
        "aggregateRating": {
            "@type": "AggregateRating",
            "bestRating": "5",
            "ratingCount": "336",
            "ratingValue": "4",
            "worstRating": "1"
        },
        "author": {
            "@type": "Person",
            "name": "Betty Crocker Kitchens"
        },
        "dateCreated": "10/12/2001",
        "description": "With just twenty minutes of prep time in the morning, you can set yourself up to come home to the inviting fragrance—not to mention flavor—of home-cooked crockpot chili. The slow-cooking perfectly combines the beef, beans and tomatoes for hearty, satisfying bowls of warm chili goodness.",
        "image": "https://images-gmi-pmc.edge-generalmills.com/f1838708-4a5c-44c6-8eae-fba0ac01197a.jpg",
        "keywords": "slow-cooker family-favorite chili",
        "name": "Slow-Cooker Family-Favorite Chili",
        "nutrition": {
            "@type": "NutritionInformation",
            "calories": "300 ",
            "carbohydrateContent": "20 g",
            "cholesterolContent": "70 mg",
            "fatContent": "0 ",
            "fiberContent": "5 g",
            "proteinContent": "25 g",
            "saturatedFatContent": "5 g",
            "servingSize": "1 1/4 cups",
            "sodiumContent": "1120 mg",
            "sugarContent": "7 g",
            "transFatContent": "1 g"
        },
        "prepTime": "PT0H20M",
        "recipeCategory": "Entree",
        "recipeIngredient": [
            "2  lb lean (at least 80%) ground beef",
            "1  large onion, chopped (1 cup)",
            "2  cloves garlic, finely chopped",
            "1  can (28 oz) Muir Glen™ organic diced tomatoes",
            "1  can (16 oz) chili beans in sauce, undrained",
            "1  can (15 oz) Muir Glen™ organic tomato sauce",
            "2  tablespoons chili powder",
            "1 1/2  teaspoons ground cumin",
            "1/2  teaspoon salt",
            "1/2  teaspoon pepper"
        ],
        "recipeInstructions": [
            {
                "@type": "HowToStep",
                "text": "In 12-inch skillet, cook beef and onion over medium heat 8 to 10 minutes, stirring occasionally, until beef is brown; drain."
            },
            {
                "@type": "HowToStep",
                "text": "In 4- to 5-quart slow cooker, mix beef mixture and remaining ingredients."
            },
            {
                "@type": "HowToStep",
                "text": "Cover and cook on Low heat setting 6 to 8 hours."
            }
        ],
        "recipeYield": "8",
        "totalTime": "PT6H20M"
    }
]
```

If multiple recipe blocks are present in the provided URL, all are extracted and displayed

## Docker Image

*docker must be installed on your local machine for the following sections to work.*

The docker commands for building a docker image are in build/package/Dockerfile.

Because Golang is a compiled language, the build can create a SCRATCH image which only contains the `scraper` binary. A SCRATCH image keeps the memory usage to the needs of the application plus a small overhead for the container runtime itself.

### Building the Docker Image

To build the docker image, from the top level directory enter the shell command:

```sh
make docker
```

The docker  image is stored in

```sh
bin/scraper.docker
```

### Running the Docker Image

From a shell enter:

```sh
docker load < bin/scraper.docker

docker run -p 8080:8080 scraper.docker
```

`scraper` is now listening on port 8080 running inside the container and will respond to the example URLs above.

