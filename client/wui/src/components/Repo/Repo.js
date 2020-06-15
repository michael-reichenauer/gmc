import React from "react";
import TableContainer from "@material-ui/core/TableContainer";
import Table from "@material-ui/core/Table";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";
import makeStyles from "@material-ui/core/styles/makeStyles";
import Typography from "@material-ui/core/Typography";
import {useSelector} from "react-redux";
import {Graph, graphWidth, rowHeight} from "../graph/Graph";

const fontSize = 13

export const useTableStyles = makeStyles((theme) => ({
    table: {
        minWidth: 650,
    },
    graphColumn: {width: 200},
    tableRow: {
        height: rowHeight,
    },
    subjectColumn: {
        width: "auto",
        border: "none",
    },
    authorColumn: {
        width: 220,
        border: "none",
    },
    timeColumn: {
        width: 130,
        border: "none",
    },
    subjectText: {
        fontSize: fontSize,
    },
    greyText: {
        color: "grey",
        fontSize: fontSize,
    },

}));

export const Repo = props => {
    const classes = useTableStyles();
    const repo = useSelector(state => state.repo)

    const spanHeight = repo.commits.length + 1
    return (
        <TableContainer className={classes.container}>
            <Table className={classes.table} size="small" padding="none">
                <TableBody>
                    <TableRow>
                        <TableCell rowSpan={spanHeight} style={{width:graphWidth(repo)}}>
                            <Graph repo={ repo} />
                        </TableCell>
                    </TableRow>
                    {repo.commits.map((commit, index) => (
                        <TableRow key={index} hover={true} className={classes.tableRow}>
                            <TableCell align={"left"} className={classes.subjectColumn}>
                                <Typography className={classes.subjectText}>{commit.subject}</Typography>
                            </TableCell>
                            <TableCell align={"left"} className={classes.authorColumn}>
                                <Typography className={classes.greyText}>{commit.author}</Typography>
                            </TableCell>
                            <TableCell align={"left"} className={classes.timeColumn}>
                                <Typography className={classes.greyText}>{commit.datetime}</Typography>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </TableContainer>
    )
}