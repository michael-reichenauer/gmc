import React from "react";
import TableContainer from "@material-ui/core/TableContainer";
import Table from "@material-ui/core/Table";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";
import makeStyles from "@material-ui/core/styles/makeStyles";
import Typography from "@material-ui/core/Typography";
import {useSelector} from "react-redux";
import {Graph} from "../graph/Graph";



export const useTableStyles = makeStyles((theme) => ({
    table: {
        minWidth: 650,
    },
    graphColumn: {width: 200},
    tableRow: {},
    subjectColumn: {width: "auto"},
    timeColumn: {
        width: 130,
    },
    timeText: {
        color: "grey",
        fontSize:12,
    },
    authorColumn: {
        width: 220,
    }
}));

export const Repo = props => {
    const classes = useTableStyles();
    const repo = useSelector(state=>state.repo)

    return (
        <TableContainer className={classes.container}>
            <Table className={classes.table} size="small" padding="none">
                <TableBody>
                    <TableRow>
                        <TableCell rowSpan={repo.length + 1} className={classes.graphColumn}>
                            <Graph width={200} height={repo.length*20}/>
                        </TableCell>
                    </TableRow>
                    {repo.map((commit, index) => (
                        <TableRow key={index} hover={true} className={classes.tableRow}>
                            <TableCell align={"left"} className={classes.subjectColumn}>
                                {commit.subject}
                            </TableCell>
                            <TableCell align={"left"} className={classes.authorColumn}>
                                <Typography className={classes.timeText}>{commit.author}</Typography>
                            </TableCell>
                            <TableCell align={"left"} className={classes.timeColumn}>
                                <Typography className={classes.timeText}>{commit.datetime}</Typography>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </TableContainer>
    )
}