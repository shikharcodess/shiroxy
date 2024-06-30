const data = [
    {
        "tags": [
            "company",
            "hot"
        ]
    },
    {
        "tags": [
            "company",
            "girl",
            "best"
        ]
    },
    {
        "tags": [
            "hot",
            "girls"
        ]
    }
]

const allTags = data.reduce((accumulator, currentValue) => {
    return accumulator.concat(currentValue.tags);
}, []);
const newSet = new Set(allTags)
const newArray = [...newSet];

console.log(newArray);
