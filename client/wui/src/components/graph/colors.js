export const masterColor = '#911eb4'
export const multiBranchColor = '#ffffff'

const branchColors = [
    '#e6194b',
    '#3cb44b',
    '#ffe119',
    '#4363d8',
    '#f58231',
    '#46f0f0',
    '#f032e6',
    '#bcf60c',
    '#fabebe',
    '#008080',
    '#e6beff',
    '#9a6324',
    '#800000',
    '#aaffc3',
    '#808000',
    '#ffd8b1',
    '#000075',
]

export function branchColor(name) {
    const index = textHash(name) % branchColors.length
    return branchColors[index]
}

function textHash(name) {
    let hash = 0, i, chr;
    for (i = 0; i < name.length; i++) {
        chr = name.charCodeAt(i);
        hash = ((hash << 5) - hash) + chr;
        hash |= 0; // Convert to 32bit integer
    }
    return hash;
}