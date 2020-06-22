// App.js
import React from "react";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import {ApplicationBar} from "./components/ApplicationBar";
import {Repo} from "./components/Repo/Repo";
import Paper from "@material-ui/core/Paper";
import {useTheme} from "./theme";
import {SnackbarProvider} from 'notistack';
import {Connection} from "./components/Connection/Connection";


export function App() {
    const [theme] = useTheme()

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
                    <MainView/>
                </SnackbarProvider>
            </Paper>
        </ThemeProvider>
    );
}


const MainView = props => {
    return (
        <>
            <ApplicationBar/>
            <Connection/>
            <Repo/>
        </>
    );
}

// useEffect(() => {
//     dispatch(connectionSlice.actions.setConnecting({isConntectin }))
//     rpcClient.Connect()
//         .then(()=>{
//             dispatch(connectingSlice.actions.set(""))
//         })
//         .catch(err =>{console.warn("Failed to connect", err)})
// },[dispatch, rpcClient]);
