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
    return arr
}

const switchToSimple = function (e, place) {

    // copy the value to the simpleInput
    simpleInput.val(place.name);

    // hide acSpan, show the simpleSpan and focus the simpleInput
    acSpan.css("display", "none");
    simpleSpan.css("display", "inline-block");
    simpleInput.focus();
}

const switchToAutocomplete = function (e, datum) {

    // copy the value to the acInput
    acInput.typeahead('val', simpleInput.val());

    // hide simpleSpan, show the acSpan and focus the acInput
    simpleSpan.css("display", "none");
    acSpan.css("display", "inline-block");
    acInput.focus();
}

// init Bloodhound
let colors_suggestions = new Bloodhound({
    // what part of the results (returned from query) to consider in Bloodhound
    datumTokenizer: function(datum) {
        return Bloodhound.tokenizers.obj.whitespace(datum.place.name);
    },

    // how to split input to be fed to the query
    queryTokenizer: Bloodhound.tokenizers.whitespace,

    // the remote configuration
    remote: {
        url: 'http://localhost:8080/api/complete?text=%QUERY',
        wildcard: '%QUERY',
        rateLimitWait: 100,

        // what to do with results before they are fed to Bloodhound
        filter: annotateDuplicates
    }
});

const myTypeaheadOptions = {
    hint: true,
    highlight: true,
    minLength: 1
}
const myTypeaheadDatasets = {
    name: 'colors',

    // what to show in the input field
    display: function(item){return item.place.name},

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
}



let completedValue;
let selectedValue;

const simpleInputInputHandler = function(e) {
    // if we backspace into the completion, re-enable the typeahead
    const newValue = simpleInput.val();

    if (newValue.length < completedValue.name.length) {
        selectedValue = null;
        completedValue = null;
        showResult();
        switchToAutocomplete()
    } else {
        if (selectedValue !== null) {
            selectedValue = null;
            showResult();
        }
        console.log(completedValue.name + " vs " + newValue)
    }
}

// queryHouseNumber queries for a house number given a placeID of a street.
const queryHouseNumber = function(placeID, houseNumber) {
    const url = 'http://localhost:8080/api/place/' + placeID + '?' + new URLSearchParams({'houseNumber': houseNumber});
    console.log(url);
    fetch(url)
        .then(response => {
            if (!response.ok) {
                 throw new Error('Network response was not OK');
            }
            return response.json();
        })
        .then(place => {
            //console.log('Success:', data)
            //completedValue = data.place;
            selectedValue = place;
            console.log("new selected", selectedValue)
            showResult();
        })
        .catch(error => {
            //console.error('Error:', error);
            return true
        });
}

const simpleInputKeypressHandler = function (e) {
    if(e.which === 13){
        const newValue = simpleInput.val();
        const houseNumber = newValue.substr(completedValue.name.length).trim()
        console.log(newValue)
        console.log(houseNumber)
        queryHouseNumber(completedValue.id, houseNumber)
    }
}




// initialize the auto complete input (acInput) -- visible
const acInput = $('#acInput')
acInput.typeahead(myTypeaheadOptions, myTypeaheadDatasets)
//acInput.on("typeahead:selected", switchToSimple);
acInput.on("typeahead:selected", function (e, datum) {
        const p = datum.place;
        completedValue = p;
        selectedValue = p;
        showResult();
        if (p.class === 'street') {
            switchToSimple(e, p);
        }
    });

const showResult = function() {
    let p = selectedValue ? selectedValue : completedValue

    if (p === null) {

        // if nothing selected, nothing to show
        document.getElementById('result').innerHTML = '&nbsp;';
    } else {
        let str;
        switch (p.class) {
            case 'street':
                str = p.name + ', ' + p.postcode + ', ' + p.district + ' (<a href="' + p.osm + '">osm</a>)'
                break;
            case 'location':
                str = p.name + ', ' + p.street + " " + p.houseNumber + ', ' + p.postcode + ', ' + p.district + ' (<a href="' + p.osm + '">osm</a>)'
            default: // 'houseNumber'
                str = p.street + " " + p.houseNumber + ', ' + p.postcode + ', ' + p.district + ' (<a href="' + p.osm + '">osm</a>)'
        }

        document.getElementById('result').innerHTML = str;
    }
}

const acSpan = $('#input span:first-child')

// initialize the simple input (simpleInput) -- hidden
const simpleInput = $('#simpleInput')
const simpleSpan = $('#simpleSpan')
simpleInput.on('input', simpleInputInputHandler);
simpleInput.on('keypress', simpleInputKeypressHandler);

// focus the acInput as document is ready
$(document).ready(function(){
    acInput.focus();
})
