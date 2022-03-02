const annotateDuplicates = (arr) => {

    let u = new Set();
    let d = new Set();
    for (const i in arr) {
        if (!u.has(arr[i].place.Name)) {
            u.add(arr[i].place.Name)
        } else {
            d.add(arr[i].place.Name)
        }
    }
    for (const i in arr) {

        if (d.has(arr[i].place.Name)) {
            arr[i].desc = arr[i].place.Postcode
        } else {
            arr[i].desc = ''
        }
    }
    return arr
}

// init Bloodhound
let colors_suggestions = new Bloodhound({
    datumTokenizer: function(datum) {
        return Bloodhound.tokenizers.whitespace(datum.place.Name);
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
        displayKey: 'Name',
        //displayKey: 'Name',
        source: colors_suggestions.ttAdapter(),
        limit: 'Infinity', // let the server descide the numbert of hits
        templates: {
            suggestion: function(data) {
                let str = '<div>';
                switch (data.place.Class)
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
                str += data.place.Name
                if (data.desc != '') {
                    str += '<span class="sugestion-discriminator">(' + data.desc + ')</span>';
                }
                return str + '</div>';
            }
        },
    })
    .on("typeahead:selected", function (e, datum) {
        console.log(datum)
    });

