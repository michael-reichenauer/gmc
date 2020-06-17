import {branchColor} from "../Graph/colors";
import {Timing} from "../../utils";

const commitCount = 500
const branchCount = 5
const mergeCount = 30

//The maximum is inclusive and the minimum is inclusive
function random(min, max) {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min + 1)) + min;
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

function merge(branchName, branchIndex1, firstCommitIndex, branchIndex2, lastCommitIndex, isBranch) {
    return {branchName, branchIndex1, firstCommitIndex, branchIndex2, lastCommitIndex, isBranch}
}


export const mockRepo = createMockData()


function createMockData() {
    const st = performance.now();
    const branches = [...Array(branchCount).keys()]
        .map(i => {
            const first = random(0, 50);
            const last = first + 1 + random(0, 50)
            return branch(i, "branch" + i, first, last);
        })
    const commits = [...Array(commitCount).keys()]
        .map(i => commit(i, "msg" + i, "michael reichenauer", '2020-06-14 ' + i))
    const merges = [...Array(mergeCount).keys()]
        .map(i => merge("msg" +  random(0, branchCount), random(0, branchCount),random(0, 10), random(0, branchCount),random(0, 10), 0 === random(0, 5) ))
    console.log("mock data: ", Timing(st))
    return {commits, branches, merges}
}
