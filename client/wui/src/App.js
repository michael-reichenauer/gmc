// App.js
import React, {useEffect, useState} from "react";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import {ApplicationBar} from "./components/ApplicationBar";
import {Repo} from "./components/Repo/Repo";
import Paper from "@material-ui/core/Paper";
import {useTheme} from "./theme";
import {Api} from "./api/api";
import {RpcClient} from "./api/RpcClient";
import {useDispatch} from "react-redux";
import {connectingSlice} from "./components/Connecting/connectingSlice";

const App = () => {
    const target =  "localhost"
    const url =  `ws://${target}:9090/api/ws`
    const rpcClient = new RpcClient(url, "api")
    const api = new Api(rpcClient)
    const dispatch = useDispatch()

    useEffect(() => {
        dispatch(connectingSlice.actions.set(target))
        rpcClient.Connect()
            .then(()=>{
                dispatch(connectingSlice.actions.set(""))
            })
            .catch(err =>{console.warn("Failed to connect", err)})
    },[dispatch, rpcClient]);

    const [value] = useTheme()

    return (
        <ThemeProvider theme={value}>
            <Paper style={{height: "100vh"}}>
                <ApplicationBar api={api}/>
                <Repo/>
            </Paper>
        </ThemeProvider>
    );
}

export default App;