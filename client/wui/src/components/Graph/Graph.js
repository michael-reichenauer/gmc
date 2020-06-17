import React from "react";
import {Circle, Layer, Line, Stage} from 'react-konva';
import {branchColor} from "./colors";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";


const columnWidth = 20
const leftMargin = 10
const rightMargin = 10
export const rowHeight = 20
const middle = rowHeight / 2
const commitSize = 3
const mergeSize = 5
const branchLineWidth = 2
const mergeLineWidth = 1


export const graphWidth = repo => repo.branches.length * columnWidth + leftMargin  +rightMargin
export const graphHeight = repo =>35 * rowHeight

export const GraphRows = props =>{
    const {repo} = props
    const spanHeight = repo.commits.length + 1

  return(
      <TableRow>
          <TableCell rowSpan={spanHeight} style={{verticalAlign: "top", width: graphWidth(repo)}}>
              <Graph repo={repo}/>
          </TableCell>
      </TableRow>
  )
}

const Graph = props => {
    const {commits, branches, merges} = props.repo
    const width = graphWidth(props.repo)
    const height = graphHeight(props.repo)

    return (
        <Stage width={width} height={height}>
            <Layer>
                <MergeLines merges = {merges}/>
                <BranchLines branches={branches}/>
                <CommitMarks commits={commits}/>
            </Layer>
        </Stage>
    );
}

function BranchLines({branches}) {
    return (
        <>
            {branches.map((b, i) =>
                <BranchLine
                    key={i}
                    index={b.index}
                    firstIndex={b.firstIndex}
                    lastIndex={b.lastIndex}
                    color={branchColor(b.name)}/>
            )}
        </>
    )
}

function MergeLines({merges}) {
    return (
        <>
            {merges.map((m, i) =>
                <MergeLine
                    key={i}
                    branchIndex1={m.branchIndex1}
                    firstIndex={m.firstCommitIndex}
                    branchIndex2={m.branchIndex2}
                    lastIndex={m.lastCommitIndex}
                    isBranch={m.isBranch}
                    color={branchColor(m.branchName)}/>
            )}
        </>
    )
}


function CommitMarks({commits}) {
    return (
        <>
            {commits.map((c, i) =>
                <CommitMark
                    key={i}
                    index={c.index}
                    branchIndex={c.branchIndex}
                    isMerge={c.isMerge}
                    color={branchColor(c.branchName)}/>
            )}
        </>
    )
}

function CommitMark({index, branchIndex, isMerge, color}) {
    const x = branchIndex * columnWidth + leftMargin
    const y = index * rowHeight + middle
    const radius = isMerge ? mergeSize : commitSize
    return (
        <Circle x={x} y={y} radius={radius} fill={color}/>
    )
}

function BranchLine({index, firstIndex, lastIndex, color}) {
    const x = index * columnWidth + leftMargin
    const y1 = firstIndex * rowHeight + middle
    const y2 = lastIndex * rowHeight + middle
    return (
        <Line points={[x, y1, x, y2]} strokeWidth={branchLineWidth} stroke={color}/>
    )
}

function MergeLine({branchIndex1, firstIndex, branchIndex2, lastIndex, isBranch, color}) {
    const x1 = branchIndex1 * columnWidth + leftMargin
    const x2 = branchIndex2 * columnWidth + leftMargin
    const y1 = firstIndex * rowHeight + middle
    const y2 = lastIndex * rowHeight + middle
  const  strokeWidth =  isBranch? branchLineWidth:mergeLineWidth
    return (
        <Line points={[x1, y1, x2, y2]} strokeWidth={strokeWidth} stroke={color}/>
    )
}