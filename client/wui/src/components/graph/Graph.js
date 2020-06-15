import React from "react";
import {Circle, Layer, Line, Stage} from 'react-konva';


const columnWidth = 20
const leftMargin = 10
export const rowHeight = 20
const middle = rowHeight / 2
const commitSize = 3
const mergeSize = 5
const branchLineWidth = 2
const mergeLineWidth = 1


export const graphWidth = repo => repo.branches.length * columnWidth + leftMargin
export const graphHeight = repo => repo.commits.length * rowHeight

export const Graph = props => {
    const {commits, branches} = props.repo
    const width = graphWidth(props.repo)
    const height = graphHeight(props.repo)
    return (
        <Stage width={width} height={height}>
            <Layer>
                <BranchLines branches={branches}/>
                <CommitMarks commits={commits}/>
            </Layer>
        </Stage>
    );
}

function BranchLines(branches) {
    return (
        <>
            <BranchLine branchIndex={2} firstIndex={2} lastIndex={50} color={0}/>
            <BranchLine branchIndex={3} firstIndex={0} lastIndex={35} color={4}/>
        </>
    )
}

// function MergeLines(merges) {
//     return (
//         <>
//             <MergeLine branchIndex1={8} firstIndex={2} branchIndex2={9} lastIndex={15} color={0}/>
//             <MergeLine branchIndex1={9} firstIndex={15} branchIndex2={8} lastIndex={35} color={4}/>
//         </>
//     )
// }


function CommitMarks({commits}) {
    return (
        <>
            {commits.map((c, i) => {
                return (
                    <CommitMark
                        key={i}
                        index={i}
                        branchIndex={3}
                        isMerge={i % 8 === 0}
                        color={4}/>
                )
            })}
        </>
    )
}

function CommitMark({index, branchIndex, isMerge, color}) {
    const x = branchIndex * columnWidth + leftMargin
    const y = index * rowHeight + middle
    const radius = isMerge ? mergeSize : commitSize
    return (
        <Circle
            x={x} y={y}
            radius={radius}
            fill={'red'}
        />
    )
}

function BranchLine({branchIndex, firstIndex, lastIndex}) {
    const x = branchIndex * columnWidth + leftMargin
    const y1 = firstIndex * rowHeight + middle
    const y2 = lastIndex * rowHeight + middle
    return (
        <Line
            points={[x, y1, x, y2]}
            strokeWidth={branchLineWidth}
            stroke={'red'}
        />
    )
}

function MergeLine({branchIndex1, firstIndex, branchIndex2, lastIndex}) {
    const x1 = branchIndex1 * columnWidth + leftMargin
    const x2 = branchIndex2 * columnWidth + leftMargin
    const y1 = firstIndex * rowHeight + middle
    const y2 = lastIndex * rowHeight + middle
    return (
        <Line
            points={[x1, y1, x2, y2]}
            strokeWidth={mergeLineWidth}
            stroke={'red'}
        />
    )
}