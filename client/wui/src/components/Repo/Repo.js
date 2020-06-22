import React, {useEffect} from "react";
import TableContainer from "@material-ui/core/TableContainer";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import makeStyles from "@material-ui/core/styles/makeStyles";
import {useDispatch, useSelector} from "react-redux";
import {GraphRows} from "../Graph/Graph";
import {CommitRows} from "../Commits/Commits";
import Typography from "@material-ui/core/Typography";
import CircularProgress from "@material-ui/core/CircularProgress";
import {repoSlice} from "./RepoSlices";
import {mockRepo} from "./mockData";
import {IsConnected, IsConnecting} from "../Connection/ConnectionSlice";


export const useTableStyles = makeStyles((theme) => ({
    table: {
        minWidth: 650,
    },
}));

export const Repo = props => {
    const classes = useTableStyles();
    const repo = useSelector(state => state.repo)
    const isConnecting = useSelector(IsConnecting)
    const isConnected = useSelector(IsConnected)
    const dispatch = useDispatch()

    useEffect(() => {
        if (isConnecting || !isConnected) {
            return
        }
        setTimeout(() => {
            dispatch(repoSlice.actions.set(mockRepo))
        }, 3000)

    }, [isConnecting, isConnected, dispatch]);

    if (!isConnecting && !isConnected) {
        return (
            <Typography>Click to connect</Typography>
        )
    }

    if (isConnecting || repo.none) {
        return (
            <CircularProgress/>
        )
    }

    return (
        <TableContainer className={classes.container}>
            <Table className={classes.table} size="small" padding="none">
                <TableBody>
                    <GraphRows repo={repo}/>
                    <CommitRows commits={repo.commits}/>
                </TableBody>
            </Table>
        </TableContainer>
    )
}