// App.js
import React, {useState} from "react";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import {ApplicationBar} from "./components/ApplicationBar/ApplicationBar";
import {Repo} from "./components/Repo/Repo";
import Paper from "@material-ui/core/Paper";
import {darkTheme} from "./theme";
import {SnackbarProvider} from 'notistack';
import {Connection} from "./components/Connection/Connection";


export function App() {
    const [theme, ]= useState(darkTheme)

    return (
        <ThemeProvider theme={theme}>
            <Paper style={{height: "100vh", backgroundColor: "black"}}>
                <SnackbarProvider
                    maxSnack={3}
                    preventDuplicate={true}
                    anchorOrigin={{
                        vertical: 'top',
                        horizontal: 'center'
                    }}>
                    <ApplicationBar/>
                    <Connection/>
                    <Repo/>
                </SnackbarProvider>
            </Paper>
        </ThemeProvider>
    );
}

