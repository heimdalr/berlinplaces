const annotateDuplicates = (arr) => {

    let u = new Set();
    let d = new Set();
    for (const i in arr) {
        if (!u.has(arr[i].place.name)) {
            u.add(arr[i].place.name)
        } else {
            d.add(arr[i].place.name)
        }
    }
    for (const i in arr) {

        if (d.has(arr[i].place.name)) {
            arr[i].desc = arr[i].place.postcode
        } else {
            arr[i].desc = ''
        }
    }
    return arr
}

// init Bloodhound
let colors_suggestions = new Bloodhound({
    datumTokenizer: function(datum) {
        return Bloodhound.tokenizers.whitespace(datum.place.name);
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
                str += data.place.name
                if (data.desc !== '') {
                    str += '<span class="sugestion-discriminator">(' + data.desc + ')</span>';
                }
                return str + '</div>';
            }
        },
    })
    .on("typeahead:selected", function (e, datum) {
        console.log(datum)
    });


$(document).ready(function(){
    document.getElementById('my_search').focus();

})
