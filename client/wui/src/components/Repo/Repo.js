import React from "react";
import TableContainer from "@material-ui/core/TableContainer";
import Table from "@material-ui/core/Table";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";
import makeStyles from "@material-ui/core/styles/makeStyles";
import {useSelector} from "react-redux";
import {GraphRows} from "../Graph/Graph";
import {Timing} from "../../utils";
import {CommitRows} from "../Commits/Commits";


export const useTableStyles = makeStyles((theme) => ({
    table: {
        minWidth: 650,
    },
}));

export const Repo = props => {
    const classes = useTableStyles();
    const repo = useSelector(state => state.repo)

    const st = performance.now()
    const r = (
        <TableContainer className={classes.container}>
            <Table className={classes.table} size="small" padding="none">
                <TableBody>
                   <GraphRows repo={repo}/>
                    <CommitRows commits={repo.commits}/>
                </TableBody>
            </Table>
        </TableContainer>
    )
    console.log("Repo: ", Timing(st))
    return r
}