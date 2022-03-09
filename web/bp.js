
// annotateDuplicates adds a discriminator string to items with the same name
const annotateDuplicates = (arr) => {

    // if there array has zero or one element, there can't be duplicates
    if (arr.length <= 1) return arr;

    // initialize the discriminator array (0 no discriminator needed, 1 district needed, 2 district and postcode needed)
    let disc = Array(arr.length).fill(0)

    // compare each pair
    for (let i = 0; i < arr.length - 1; i++) {
        for (let j = i+1; j < arr.length; j++) {
            const ii = arr[i].place;
            const ij = arr[j].place;

            // if names are equal
            if (ii.name === ij.name) {

                // if districts are equal, we also need postcode to discriminate
                if (ii.district === ij.district) {
                    disc[i] = disc[j] = 2;
                } else {
                    disc[i] = disc[i] === 2 ? 2 : 1;
                    disc[j] = disc[j] === 2 ? 2 : 1;
                }
            }
        }
    }

    // set discriminator based on needs computed above
    for (let i = 0; i < arr.length; i++) {
        const ii = arr[i].place;
        switch (disc[i]) {
            case 2:
                ii.disc = ii.postcode + ', ' + ii.district
                break
            case 1:
                ii.disc = ii.district
                break
            default:
                ii.disc = ''
        }
    }

    // for (let i = 0; i < arr.length; i++) {
    //     arr[i].name = arr[i].place.name
    // }
    return arr
}

// init Bloodhound
let colors_suggestions = new Bloodhound({
    datumTokenizer: function(datum) {
        return Bloodhound.tokenizers.whitespace(datum.place.name);
    },
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    remote: {
        url: 'http://localhost:8080/api/complete?text=%QUERY',
        wildcard: '%QUERY',
        rateLimitWait: 100,
        filter: annotateDuplicates
      }
});

// init Typeahead
$('#acInput').typeahead({
        hint: true,
        highlight: true,
        minLength: 1
    },
    {
        name: 'colors',
        displayKey: 'name',
        source: colors_suggestions.ttAdapter(),
        limit: 'Infinity', // let the server descide the number of hits
        templates: {
            suggestion: function(data) {
                let str = '<div>';
                if (data.place.class === 'location') {
                    if (['house', 'museum', 'hostel', 'hotel'].includes(data.place.type)) {
                        str += '<i class="bi bi-bank"></i>'
                    } else if (['nightclub', 'pub', 'restaurant', 'house', 'cafe', 'biergarten', 'bar'].includes(data.place.type)) {
                        str += '<i class="bi bi-cup-straw"></i>'
                    }
                } else {
                        str += '<i class="bi bi-geo-alt"></i>'
                }
                str += data.place.name
                if (data.place.disc !== '') {
                    str += '<span class="sugestion-discriminator">(' + data.place.disc + ')</span>';
                }
                return str + '</div>';
            }
        },
    })
    .on("typeahead:selected", function (e, datum) {
        const p = datum.place;
        let str;
        switch (p.class)
        {
            case 'street':
                str = p.name + ' ' + p.postcode + ', ' + p.district  + ' (<a href="' + p.osm + '">osm</a>)'
                break;
            case 'location':
                str = p.name + ' ' + p.street + " " + p.houseNumber + ', ' + p.postcode + ', ' + p.district  + ' (<a href="' + p.osm + '">osm</a>)'
            default: // 'houseNumber'
                str = p.street + " " + p.houseNumber + ', ' + p.postcode + ', ' + p.district  + ' (<a href="' + p.osm + '">osm</a>)'
        }

        document.getElementById('result').innerHTML = str;
    });


$(document).ready(function(){
    document.getElementById('acInput').focus();

})
