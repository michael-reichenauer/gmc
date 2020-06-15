import {branchColor} from "../graph/colors";

const commitCount = 1500
const branchCount = 5

function random(min, max) {
    return Math.floor(Math.random() * max) + min

}

function commit(index, subject, author, datetime) {
    const branchIndex = random(0, branchCount)
    const branchName = "branch" + branchIndex
    const isMerge = random(0, 10)===0
    return {index, subject, author, datetime, branchName, branchIndex, isMerge}
}

function branch(index, name, firstIndex, lastIndex) {
    return {index, name, firstIndex, lastIndex}
}


export const mockRepo = createMockData()


function createMockData() {
    const branches = [...Array(branchCount).keys()]
        .map(i => {
            const first = random(0, 50);
            const last = first + 1 + random(0, 50)
            return branch(i, "branch" + i, first, last);
        })
    const commits = [...Array(commitCount).keys()]
        .map(i => commit(i, "msg" + i, "michael reichenauer", '2020-06-14 ' + i))

    return {commits, branches}
}
