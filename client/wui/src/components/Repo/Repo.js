import React from "react";
import TableContainer from "@material-ui/core/TableContainer";
import Paper from "@material-ui/core/Paper";
import Table from "@material-ui/core/Table";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";
import rows from "./mockData";
import makeStyles from "@material-ui/core/styles/makeStyles";


export const useTableStyles = makeStyles((theme) => ({
    table: {
        minWidth: 650,
    },
}));

export const Repo = props => {
    const classes = useTableStyles();

    return (
        <TableContainer component={Paper}>
            <Table className={classes.table} size="small">
                <TableBody>
                    <TableRow>
                        <TableCell rowSpan={rows.length + 1}>
                            Graph
                        </TableCell>
                    </TableRow>
                    {rows.map((row, index) => (
                        <TableRow key={index}>
                            <TableCell align={"left"}>{row.subject}</TableCell>
                            <TableCell align={"left"}>{row.author}</TableCell>
                            <TableCell align={"left"}>{row.datetime}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </TableContainer>
    )
}