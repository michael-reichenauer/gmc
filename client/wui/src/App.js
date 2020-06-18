// App.js
import React, {useEffect} from "react";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import {ApplicationBar} from "./components/ApplicationBar";
import {Repo} from "./components/Repo/Repo";
import Paper from "@material-ui/core/Paper";
import {useTheme} from "./theme";
import {Api} from "./api/api";

const App = () => {
    useEffect(() => {
        const api = new Api("ws://localhost:8080/ws",
            msg=>{console.log("Received message: ", msg)})
        api.connect()
        console.log("Called connect");
    },[]);

    const [value] = useTheme()
    return (
        <ThemeProvider theme={value}>
            <Paper style={{height: "100vh"}}>
                <ApplicationBar/>
                <Repo/>
            </Paper>
        </ThemeProvider>
    );
}

export default App;