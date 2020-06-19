// App.js
import React, {useEffect} from "react";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import {ApplicationBar} from "./components/ApplicationBar";
import {Repo} from "./components/Repo/Repo";
import Paper from "@material-ui/core/Paper";
import {useTheme} from "./theme";
import {Api} from "./api/api";
import {RpcClient} from "./api/RpcClient";

const App = () => {
    const rpcClient = new RpcClient("ws://localhost:9090/api/ws", "api")
    const api = new Api(rpcClient)
    useEffect(() => {
        console.log("Called connect");
        rpcClient.Connect()
            .then(()=>{console.log("Connected")})
            .catch(err =>{console.warn("Failed to connect", err)})
    },[rpcClient]);

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