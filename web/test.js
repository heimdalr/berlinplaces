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

const myTypeaheadOptions = {
    hint: true,
    highlight: true,
    minLength: 1
}
const myTypeaheadDatasets = {
    source: colors_suggestions.ttAdapter(),
    limit: 'Infinity',
}

let selectedValue;

const simpleInputInputHandler = function(e) {
    // if we backspace into the selection, re-enable the typeahead
    const newValue = simpleInput.val();
    if (newValue.length < selectedValue.length) {
        switchToAutocomplete()
    } else {
        console.log(selectedValue + " vs " + newValue)
    }
}

const simpleInputKeypressHandler = function (e) {
    if(e.which === 13){
        const newValue = simpleInput.val();
        console.log(newValue)
    }
}

const switchToSimple = function (e, datum) {

    selectedValue = datum;

    // display the selected value above the input
    // document.getElementById('result').innerHTML = datum;

    // copy the value to the simpleInput
    simpleInput.val(selectedValue);

    // hide acSpan, show the simpleSpan and focus the simpleInput
    acSpan.css("display", "none");
    simpleSpan.css("display", "inline-block");
    simpleInput.focus();
}

const switchToAutocomplete = function (e, datum) {

    const simpleValue = simpleInput.val()

    // copy the value to the acInput
    acInput.typeahead('val', simpleValue);

    // hide simpleSpan, show the acSpan and focus the acInput
    simpleSpan.css("display", "none");
    acSpan.css("display", "inline-block");
    acInput.focus();
}


// initialize the auto complete input (acInput) -- visible
const acInput = $('#acInput')
acInput.typeahead(myTypeaheadOptions, myTypeaheadDatasets)
acInput.on("typeahead:selected", switchToSimple);
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
