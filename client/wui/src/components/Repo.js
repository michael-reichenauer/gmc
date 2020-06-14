import React from "react";
import TableContainer from "@material-ui/core/TableContainer";
import Paper from "@material-ui/core/Paper";
import Table from "@material-ui/core/Table";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";
import rows from "../mockData";
import makeStyles from "@material-ui/core/styles/makeStyles";


export const useTableStyles = makeStyles((theme) => ({
    table: {
        minWidth: 650,
    },
}));

export const Repo = props =>{
    const classes = useTableStyles();

    return(
        <TableContainer component={Paper}>
            <Table className={classes.table} size="small">
                <TableHead>
                    <TableRow>
                        <TableCell>Dessert</TableCell>
                        <TableCell align="right">Calories</TableCell>
                        <TableCell align="right">Fat</TableCell>
                        <TableCell align="right">Carbs</TableCell>
                        <TableCell align="right">Protein</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {rows.map((row, index) => (
                        <TableRow key={index}>
                            <TableCell component="th" scope="row">{row.name}</TableCell>
                            <TableCell align={"right"}>{row.calories}</TableCell>
                            <TableCell align={"right"}>{row.fat}</TableCell>
                            <TableCell align={"right"}>{row.carbs}</TableCell>
                            <TableCell align={"right"}>{row.protein}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </TableContainer>
    )
}