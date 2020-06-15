// App.js
import React from "react";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import {ApplicationBar} from "./components/ApplicationBar";
import {Repo} from "./components/Repo/Repo";
import Paper from "@material-ui/core/Paper";
import {useTheme} from "./theme";

const App = () => {
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