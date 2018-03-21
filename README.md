
Beermap is a custom map server based on google maps.

A map is built from a list text file, an icon style file and an optional google maps style.

# List text file should be in format:

	[iconlabel1] name1
	(address_or_openlocationcode_or_latlong)
	#tag1 #tag2
	text

	[iconlabel2] name2
	(address_or_openlocationcode_or_latlong)
	#tag1 #tag2
	text

Example:

	[1] Eiffel Tower, Paris
	(V75V+8Q Paris, France)
	#visited

	[2] Tower Bridge, London
	(Tower Bridge, London)

TODO: use name if address is missing.

# Icon style file

Icon style is a json for rendering icons. The specified font must be available from Google fonts.

TODO: add font option to individual styles and additional font sources.
TODO: add glow/shadow settings, especially for dark maps.

Example:

	{
		"font": "Roboto:500",
		"styles": [{
			"name": "closed",
			"cond":{
				type: "tag",
				value: "#closed"
			},
			"color": "#b22222",
			"shape":"circle"
		}, {
			"name": "visited",
			"cond":{
				type: "tag",
				value: "#visited"
			},
			"color": "#228b22",
			"shape":"circle"
		}, {
			"name": "hotel",
			"cond":{
				type: "tag",
				value: "#hotel"
			},
			"color": "#878401",
			"shape":"square"
		}, {
			"name": "other",
			"color": "#1e90ff",
			"shape":"circle"
		}]
	}

# Map style file

Map style is for the google map UI. A nice source of styles is [snazzy maps](https://snazzymaps.com/).

