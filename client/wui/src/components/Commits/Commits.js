import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import Typography from "@material-ui/core/Typography";
import React from "react";
import makeStyles from "@material-ui/core/styles/makeStyles";
import {rowHeight} from "../Graph/Graph";


const fontSize = 13
export const useTableStyles = makeStyles((theme) => ({

    tableRow: {
        height: rowHeight,
    },
    subjectColumn: {
        width: "auto",
        border: "none",
    },
    subjectText: {
        fontSize: fontSize,
    },
    authorColumn: {
        width: 220,
        border: "none",
    },
    timeColumn: {
        width: 130,
        border: "none",
    },
    authorTimeText: {
        color: "grey",
        fontSize: fontSize,
    },
}));


export const CommitRows = props => {
    const {commits} = props
    const classes = useTableStyles();

    return (
        <>
            {commits.map((commit, index) => (
                <TableRow key={index} hover={true} className={classes.tableRow}>
                    <TableCell align={"left"} className={classes.subjectColumn}>
                        <Typography className={classes.subjectText}>{commit.subject}</Typography>
                    </TableCell>
                    <TableCell align={"left"} className={classes.authorColumn}>
                        <Typography className={classes.authorTimeText}>{commit.author}</Typography>
                    </TableCell>
                    <TableCell align={"left"} className={classes.timeColumn}>
                        <Typography className={classes.authorTimeText}>{commit.datetime}</Typography>
                    </TableCell>
                </TableRow>
            ))}
        </>
    )
}