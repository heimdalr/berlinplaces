// completedValue tracks places identified through autocomplete.
let completedValue;

// selectedValue tracks refined places (i.e. in case of an autocompleted street, selectedValue tracks house numbers).
let selectedValue;

// showResult displays the autocompleted / selected address above the simple input field or typeahead input field.
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
                break;
            default: // 'houseNumber'
                str = p.street + " " + p.houseNumber + ', ' + p.postcode + ', ' + p.district + ' (<a href="' + p.osm + '">osm</a>)'
        }

        document.getElementById('result').innerHTML = str;
    }
}

// annotateDuplicates adds a discriminator string to items with the same name.
const annotateDuplicates = (arr) => {

    // if the array has less than two elements, there can't be duplicates
    if (arr.length < 2) return arr;

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

    // set discriminator string based on the discriminator array computed above
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

// switchToSimple switches to the simple input field (i.e. no typeahead).
const switchToSimple = function () {

    // copy value from typeahead input to the simple input
    simpleInput.val(acInput.typeahead('val'));

    // hide acSpan, show the simpleSpan and focus the simpleInput
    acSpan.css("display", "none");
    simpleSpan.css("display", "inline-block");
    simpleInput.focus();
}

// switchToTypeahead switches (back) to the typeahead input field.
const switchToTypeahead = function () {

    // copy value from simple input to typeahead input
    acInput.typeahead('val', simpleInput.val());

    // hide simpleSpan, show the acSpan and focus the acInput
    simpleSpan.css("display", "none");
    acSpan.css("display", "inline-block");
    acInput.focus();
}

// myBloodhoundConfiguration is the Bloodhound config.
const myBloodhoundConfiguration = new Bloodhound({

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
        rateLimitWait: 10,

        // what to do with results before they are fed to Bloodhound
        filter: annotateDuplicates
    }
});

// myTypeaheadOptions are the Typeahead Options.
const myTypeaheadOptions = {
    hint: true,
    highlight: true,
    minLength: 1
}

// myTypeaheadDataset is the Typeahead (remote) Dataset (using the myBloodhoundConfiguration).
const myTypeaheadDataset = {
    name: 'colors',

    // what to show in the input field
    display: function(item){return item.place.name},

    source: myBloodhoundConfiguration.ttAdapter(),
    limit: 'Infinity', // let the server decide the number of hits
    templates: {

        // how to display suggestions (i.e. list items in the Typeahead dropdown menu)
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

// simpleInputChangeHandler is the handler for changes in the simple input.
const simpleInputChangeHandler = function() {

    const newValue = simpleInput.val();

    // if we backspace into the completion
    if (newValue.length < completedValue.name.length) {

        // re-enable the typeahead
        selectedValue = null;
        completedValue = null;
        showResult();
        switchToTypeahead()
    } else {

        // erase selected value if necessary
        if (selectedValue !== null) {
            selectedValue = null;
            showResult();
        }
    }
}

// simpleInputKeypressHandler is the handler for Return- / Enter-key events in the simple input.
const simpleInputKeypressHandler = function (e) {
    if(e.which === 13){

        // compute the part of input value that does not belong to the autocompleted value (i.e. the house number)
        const newValue = simpleInput.val();
        const houseNumber = newValue.substr(completedValue.name.length).trim()

        // queries for a house number given the id of a street
        const url = 'http://localhost:8080/api/place/' + completedValue.id + '?' + new URLSearchParams({'houseNumber': houseNumber});
        fetch(url)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not OK');
                }
                return response.json();
            })
            .then(place => {
                selectedValue = place;
                showResult();
            })
            .catch(_ => {
                return true
            });
    }
}

// typeaheadSelected is the handler for selected events in the typeahead input.
const typeaheadSelected = function (_, datum) {
    const p = datum.place;
    completedValue = p;
    selectedValue = p;
    showResult();

    // if selected place is a street switch to the simple input to allow entering and selecting house numbers
    if (p.class === 'street') {
        switchToSimple();
    }
}

// initialize the typeahead input (initially displayed)
const acInput = $('#acInput')
acInput.typeahead(myTypeaheadOptions, myTypeaheadDataset)
acInput.on("typeahead:selected", typeaheadSelected);
const acSpan = $('#input span:first-child')

// initialize the simple input (initially NOT displayed but substitutes the typeahead input as needed)
const simpleInput = $('#simpleInput')
const simpleSpan = $('#simpleSpan')
simpleInput.on('input', simpleInputChangeHandler);
simpleInput.on('keypress', simpleInputKeypressHandler);

// focus the typeahead input as document is ready
$(document).ready(function(){
    acInput.focus();
})
