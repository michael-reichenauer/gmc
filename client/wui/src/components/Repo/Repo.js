import React, {useEffect} from "react";
import TableContainer from "@material-ui/core/TableContainer";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import makeStyles from "@material-ui/core/styles/makeStyles";
import {GraphRows} from "../Graph/Graph";
import {CommitRows} from "../Commits/Commits";
import CircularProgress from "@material-ui/core/CircularProgress";
import {mockRepo} from "./mockData";
import {IsConnected, IsConnecting} from "../Connection/ConnectionSlice";
import {useRepo} from "./RepoState";
import {useSelector} from "react-redux";


export const useTableStyles = makeStyles((theme) => ({
    table: {
        minWidth: 650,
    },
}));

export const Repo = props => {
    const classes = useTableStyles();
    const isConnecting = useSelector(IsConnecting)
    const isConnected = useSelector(IsConnected)
    const [repo, setRepo]  = useRepo()

    useEffect(() => {
        if (isConnecting || !isConnected) {
            return
        }
        setRepo(mockRepo)
        // setTimeout(() => {
        //     setRepo(mockRepo)
        // }, 3000)

    }, [isConnecting, isConnected, setRepo]);

    if (!isConnecting && !isConnected) {
        return <div/>
    }

    if (isConnecting || repo == null) {
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