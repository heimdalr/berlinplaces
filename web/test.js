const states = ['Alabama', 'Alaska', 'Arizona', 'Arkansas', 'California',
    'Colorado', 'Connecticut', 'Delaware', 'Florida', 'Georgia', 'Hawaii',
    'Idaho', 'Illinois', 'Indiana', 'Iowa', 'Kansas', 'Kentucky', 'Louisiana',
    'Maine', 'Maryland', 'Massachusetts', 'Michigan', 'Minnesota',
    'Mississippi', 'Missouri', 'Montana', 'Nebraska', 'Nevada', 'New Hampshire',
    'New Jersey', 'New Mexico', 'New York', 'North Carolina', 'North Dakota',
    'Ohio', 'Oklahoma', 'Oregon', 'Pennsylvania', 'Rhode Island',
    'South Carolina', 'South Dakota', 'Tennessee', 'Texas', 'Utah', 'Vermont',
    'Virginia', 'Washington', 'West Virginia', 'Wisconsin', 'Wyoming'
];

// init Bloodhound
let colors_suggestions = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    // `states` is an array of state names defined in "The Basics"
    local: states
});

let selectedValue;

const myTypeaheadOptions = {
    hint: true,
    highlight: true,
    minLength: 1
}
const myTypeaheadDatasets = {
    source: colors_suggestions.ttAdapter(),
    limit: 'Infinity',
}


const inputHandler = function(e) {
    // if we backspace into the selection, re-enable the typeahead
    const newValue = $('#my_search2').val();
    if (newValue.length < selectedValue.length) {

        // update the value of the typeahead box
        $('#my_search').typeahead('val', newValue);

        // hide the typeahead input and show and focus the simple input
        $('#my_search2').css("display", "none");
        $('#my_search').css("display", "block");
        $('#my_search').focus();


    } else {
        console.log(selectedValue + " vs " + newValue)
    }
}

const enterTracker = function (e) {
    if(e.which === 13){
        const newValue = $('#my_search').val();

        // document.getElementById('result').innerHTML = newValue;
        console.log("enter pressed")
    }
}

const myTypeaheadDisable2 = function (e, datum) {

    // safe the selected value
    selectedValue = datum;

    // display the selected value above the input
    // document.getElementById('result').innerHTML = datum;

    // copy the value to our input
    $('#my_search2').val(datum);

    // hide the typeahead input and show and focus the simple input
    $('#my_search').css("display", "none");
    $('#my_search2').css("display", "block");
    $('#my_search2').focus();

    console.log(datum)


}

$('#my_search').typeahead(myTypeaheadOptions, myTypeaheadDatasets)
$('#my_search').on("typeahead:selected", myTypeaheadDisable2);

$('#my_search2').on('input', inputHandler);
$('#my_search2').on('keypress', enterTracker);

$(document).ready(function(){
    document.getElementById('my_search').focus();

})
