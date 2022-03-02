const annotateDuplicates = (arr) => {


    // compute a set of duplicate names
    let u = new Set(); // set to store names we have seen
    let d = new Set(); // set to store names that appear several times
    for (const i in arr) {

        // get the name of the current place
        const name = arr[i].place.name;

        // copy name to object root (typeahed likes it there)
        arr[i].name = name

        // if we have seen this name already add it to the duplicate set
        if (!u.has(name)) {
            u.add(name)
        } else {
            d.add(name)
        }
    }

    // for each item which with a name that appears several times in the result set (an entry in the duplicates set),
    // add the postcode as discriminator
    for (const i in arr) {
        if (d.has(arr[i].name)) {
            arr[i].disc = arr[i].place.postcode
        } else {
            arr[i].disc = ''
        }
    }

    return arr
}

// init Bloodhound
let colors_suggestions = new Bloodhound({
    datumTokenizer: function(datum) {
        return Bloodhound.tokenizers.whitespace(datum.name);
    },
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    remote: {
        url: 'http://localhost:8080/api/?text=%QUERY',
        wildcard: '%QUERY',
        rateLimitWait: 100,
        filter: annotateDuplicates
      }
});

// init Typeahead
$('#my_search').typeahead({
        hint: true,
        highlight: true,
        minLength: 1
    },
    {
        name: 'colors',
        displayKey: 'name',
        source: colors_suggestions.ttAdapter(),
        limit: 'Infinity', // let the server descide the numbert of hits
        templates: {
            suggestion: function(data) {
                let str = '<div>';
                switch (data.place.class)
                {
                    case 'amenity':
                        str += '<i class="bi bi-cup-straw"></i>'
                        break;
                    case 'tourism':
                        str += '<i class="bi bi-bank"></i>'
                        break;
                    default: // 'highway':
                        str += '<i class="bi bi-geo-alt"></i>'
                }
                str += data.name
                if (data.disc !== '') {
                    str += '<span class="sugestion-discriminator">(' + data.disc + ')</span>';
                }
                return str + '</div>';
            }
        },
    })
    .on("typeahead:selected", function (e, datum) {
        document.getElementById('result').innerHTML = datum.name + " " + datum.place.houseNumber + ", " + datum.place.postcode + " " + datum.place.city
    });


$(document).ready(function(){
    document.getElementById('my_search').focus();

})
