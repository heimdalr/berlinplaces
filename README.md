# berlinplaces


[![Test](https://github.com/heimdalr/berlinplaces/actions/workflows/test.yml/badge.svg)](https://github.com/heimdalr/berlinplaces/actions/workflows/test.yml)
<!--
[![Coverage Status](https://coveralls.io/repos/github/heimdalr/arangodag/badge.svg?branch=main)](https://coveralls.io/github/heimdalr/arangodag?branch=main)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/heimdalr/arangodag)](https://pkg.go.dev/github.com/heimdalr/arangodag)
[![Go Report Card](https://goreportcard.com/badge/github.com/heimdalr/arangodag)](https://goreportcard.com/report/github.com/heimdalr/arangodag)
-->

REST-Service for autocompletion and geocoding of places and addresses in Berlin.

berlinplaces is essentially me playing around with [Open Street Map](https://wiki.osmfoundation.org/wiki/Main_Page)
data. The goal is (was), to imitate Google's [Places Autocomplete](https://developers.google.com/maps/documentation/javascript/places-autocomplete#introduction)
(-API) without the strings attached. That is, provide an API that is free (beer and speech), has a low latency, has a 
good "hit rate" (e.g. compensates typos), and is slim and easy in terms of deployment. 

Thus here it is, berlinplaces is:

- free: it's here and OSS
- [latency](#latency): basic tests show ~200µs without typos and ~12ms with early typos (if completed first time)
- hit rate: berlinplaces uses lookup tables for speed and Levenshtein for typos
- slim and easy: 25MB Docker image (incl. REST-server, OSM-data, swagger-docs and example website) 

The demo (see below) looks like:

[![demo](berlinplaces.png)](berlinplaces.gif)
  
## Getting Started

Have [Go](https://go.dev/) >= 1.17 installed and run: 

~~~~bash
git clone git@github.com:heimdalr/berlinplaces.git
cd berlinplaces
go build -o berlinplaces .
./berlinplaces 
~~~~

and surf to:

- <http://localhost:8080/web> - demo website or
- <http://localhost:8080/swagger> - the OpenAPI spec

alternatively run (e.g.): 

~~~~bash
curl --request GET --url 'http://localhost:8080/api/?text=Oranienbur' | jq
~~~~

which will result in something like:

~~~~json
[
  {
    "distance": 1,
    "percentage": 0,
    "place": {
      "placeID": 715227,
      "parentPlaceID": 732833,
      "class": "highway",
      "type": "secondary",
      "name": "Oranienburger Straße",
      "boundary": "Wittenau",
      "postcode": "13437",
      "city": "Berlin",
      "lat": 52.5933948,
      "lon": 13.3335015,
      "relevance": 0,
      "simpleName": "oranienburgerstraße"
    }
  },
  {
    "distance": 3,
    "percentage": 0,
    "place": {
      "placeID": 523006,
      "parentPlaceID": 465333,
      "class": "highway",
      "type": "unclassified",
      "name": "Brandenburger Straße",
      "boundary": "Wedding",
      "neighbourhood": "Cité Joffre",
      "postcode": "13405",
      "city": "Berlin",
      "lat": 52.5555049,
      "lon": 13.3115449,
      "relevance": 0,
      "simpleName": "brandenburgerstraße"
    }
  }
]
~~~~

## OSM Data

The repository at hand contains prepared OSM data ([`berlin.csv`](berlin.csv)). 

See [`_data/README.md`](_data/README.md) for how to generate this CSV file.  

## Latency

In the following we look at different lookup latency based on:

- `maxPrefixLength = 6` (maximum prefix length)
- `minCompletionCount = 6` (the minimum number of completions to compute)
- `levMinimum = 0` (the minimum input length before doing Levenshtein)

Essential basic tests show ~4ms without typos ~13ms with early typos (locally, on an i5-4670S).

### Without typos

Autocompleting on "oranienburger straße":

~~~~
[GIN] | 200 | 484.152µs | GET "/api/?text=o"
[GIN] | 200 | 215.291µs | GET "/api/?text=or"
[GIN] | 200 |  218.64µs | GET "/api/?text=ora"
[GIN] | 200 | 156.601µs | GET "/api/?text=oran"
[GIN] | 200 | 147.172µs | GET "/api/?text=orani"
[GIN] | 200 | 305.857µs | GET "/api/?text=oranie"
[GIN] | 200 | 202.536µs | GET "/api/?text=oranien" --> visible
[GIN] | 200 | 270.033µs | GET "/api/?text=oranienb"
[GIN] | 200 | 170.817µs | GET "/api/?text=oranienbu"
[GIN] | 200 | 188.221µs | GET "/api/?text=oranienbur"
[GIN] | 200 | 274.064µs | GET "/api/?text=oranienburg"
[GIN] | 200 | 220.231µs | GET "/api/?text=oranienburge"
[GIN] | 200 | 220.364µs | GET "/api/?text=oranienburger"
[GIN] | 200 | 148.976µs | GET "/api/?text=oranienburger%20"
[GIN] | 200 | 151.191µs | GET "/api/?text=oranienburger%20s" --> at top
[GIN] | 200 |  200.57µs | GET "/api/?text=oranienburger%20st"
[GIN] | 200 | 292.231µs | GET "/api/?text=oranienburger%20sta"
[GIN] | 200 | 190.156µs | GET "/api/?text=oranienburger%20star"
[GIN] | 200 |  222.93µs | GET "/api/?text=oranienburger%20star%C3%9F"
[GIN] | 200 | 184.946µs | GET "/api/?text=oranienburger%20star%C3%9Fe"
~~~~

The average response time over all 20 calls (one for each character typed) is ~220µs. 

The correct "Oranienburger Straße" is suggested after typing "oranien" and at the top of the suggestion list after 
typing "oranienburger s".

### Early typos

Early typos are typos that occur inside the prefix lookup.

Autocompleting on "oanienburgerstraße" (note the missing "r" in the beginning):

~~~~
[GIN] | 200 |   422.906µs | GET "/api/?text=o"
[GIN] | 200 |   210.837µs | GET "/api/?text=oa"
[GIN] | 200 |  9.644704ms | GET "/api/?text=oan"
[GIN] | 200 |  10.18646ms | GET "/api/?text=oani"
[GIN] | 200 | 10.554832ms | GET "/api/?text=oanie"
[GIN] | 200 | 11.641434ms | GET "/api/?text=oanien"
[GIN] | 200 | 10.935276ms | GET "/api/?text=oanienb"
[GIN] | 200 | 12.162737ms | GET "/api/?text=oanienbu"
[GIN] | 200 | 12.863149ms | GET "/api/?text=oanienbur"
[GIN] | 200 | 14.113217ms | GET "/api/?text=oanienburg"
[GIN] | 200 | 12.674369ms | GET "/api/?text=oanienburge"
[GIN] | 200 | 13.981331ms | GET "/api/?text=oanienburger"
[GIN] | 200 | 14.620061ms | GET "/api/?text=oanienburgers" --> visible
[GIN] | 200 | 16.405334ms | GET "/api/?text=oanienburgerst" --> at top
[GIN] | 200 | 15.504753ms | GET "/api/?text=oanienburgerstr"
[GIN] | 200 | 19.162968ms | GET "/api/?text=oanienburgerstra"
[GIN] | 200 | 17.305315ms | GET "/api/?text=oanienburgerstra%C3%9F"
[GIN] | 200 | 19.710446ms | GET "/api/?text=oanienburgerstra%C3%9Fe"
~~~~

Early typos ruin the lookup. The average response time over all 18 calls is ~12ms. 

The correct "Oranienburger Straße" is suggested after typing "oanienburgers" and at the top of the suggestion list after
typing "oanienburgerst".

In this case, there are no prepared completion for the prefix "oa" (and following). Thus, berlinplaces does Levenshtein
on the complete set for this call and all subsequent prefixes.(see [this](https://github.com/heimdalr/berlinplaces/issues/1) 
issue.)

Now, good news, berlinplaces caches the results of input completions for faulty inputs and inputs longer than the 
configured `maxPrefixLength`. Thus running the same faulty input ("oanienburgerstraße") again, results in an average
response length of 194µs.

### Late typos

Late typos are typos that occur outside / after the prefix lookup.  

Autocompleting on "oranienurgerstarße" (note the missing "b" and the flipped "ar" vs. "ra"):

~~~~
[GIN] | 200 | 249.977µs | GET "/api/?text=o"
[GIN] | 200 | 358.533µs | GET "/api/?text=or"
[GIN] | 200 | 159.317µs | GET "/api/?text=ora"
[GIN] | 200 | 297.679µs | GET "/api/?text=oran"
[GIN] | 200 | 225.726µs | GET "/api/?text=orani"
[GIN] | 200 | 228.456µs | GET "/api/?text=oranie"
[GIN] | 200 | 162.784µs | GET "/api/?text=oranien" --> visible
[GIN] | 200 | 169.192µs | GET "/api/?text=oranienu"
[GIN] | 200 | 195.499µs | GET "/api/?text=oranienur"
[GIN] | 200 | 211.435µs | GET "/api/?text=oranienurg"
[GIN] | 200 | 138.643µs | GET "/api/?text=oranienurge"
[GIN] | 200 |  165.08µs | GET "/api/?text=oranienurger"
[GIN] | 200 | 200.702µs | GET "/api/?text=oranienurgers"
[GIN] | 200 | 182.195µs | GET "/api/?text=oranienurgerst" --> at top
[GIN] | 200 | 163.033µs | GET "/api/?text=oranienurgersta"
[GIN] | 200 | 208.039µs | GET "/api/?text=oranienurgerstar"
[GIN] | 200 | 302.667µs | GET "/api/?text=oranienurgerstar%C3%9F"
[GIN] | 200 | 162.801µs | GET "/api/?text=oranienurgerstar%C3%9Fe"
~~~~

Late typos are cheap as Levenshtein will only be done on the completion set delivered by the prefix lookup. The average 
response time over all 18 calls is in this case ~210µs (essentially the same as for no typos).

The correct "Oranienburger Straße" is suggested after typing "oranien" and at the top of the suggestion list after
typing "oranienurgerst".

