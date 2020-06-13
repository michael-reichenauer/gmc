// App.js
import React, {useState} from "react";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import {darkTheme, lightTheme} from "./theme";
import {ApplicationBar} from "./components/ApplicationBar";
import {Repo} from "./components/Repo";
import Paper from "@material-ui/core/Paper";


const App = () => {
    const [lightMode, setLightMode] = useState(false)

    return (
        <ThemeProvider theme={lightMode ? lightTheme : darkTheme}>
            <Paper style={{height: "100vh"}}>
                <ApplicationBar/>
                <Repo/>
            </Paper>
        </ThemeProvider>
    );
}

export default App;