const commitCount = 1500
const branchCount = 5

function random(min, max) {
     return Math.floor(Math.random() * branchCount) + min

}
function commit(index, subject, author, datetime) {
    const branchName= "branch" + random(0, branchCount)
    return {index, subject, author, datetime, branchName}
}

function branch(name, firstIndex, lastIndex) {
    return {name, firstIndex, lastIndex}
}


export const mockRepo =createMockData()


function createMockData()  {
    const branches =  [...Array(branchCount).keys()]
        .map(i =>{
            const first =  random(0, 50);
            const last = first +1+ random(0, 50)
            return branch("branch"+i, first, last);
        })
    const commits =  [...Array(commitCount).keys()]
        .map(i =>commit(i, "msg"+i, "michael reichenauer", '2020-06-14 '+i))
    
    return {commits, branches}
}
