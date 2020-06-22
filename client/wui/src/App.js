// App.js
import React, {useState} from "react";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import {ApplicationBar} from "./components/ApplicationBar";
import {Repo} from "./components/Repo/Repo";
import Paper from "@material-ui/core/Paper";
import {useTheme} from "./theme";
import {Api} from "./api/api";
import {RpcClient} from "./api/RpcClient";
import {useDispatch, useSelector} from "react-redux";
import {connectionSlice, IsConnected, IsConnecting} from "./components/Connection/ConnectionSlice";
import {SnackbarProvider, useSnackbar} from 'notistack';

const App = () => {
    const [value] = useTheme()

    return (
        <ThemeProvider theme={value}>
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

export default App;

const MainView = props => {
    const target = "localhost"
    const url = `ws://${target}:9090/api/ws`
    const dispatch = useDispatch()
    const {enqueueSnackbar, closeSnackbar} = useSnackbar();
    const isConnecting = useSelector(IsConnecting)
    const isConnected = useSelector(IsConnected)

    const onCloseError = err => {
        dispatch(connectionSlice.actions.setError(err))
        enqueueSnackbar(`Connection failed to ${target}`, {variant: "error", onClick: () => closeSnackbar()})
    }

    const [rpc] = useState(new RpcClient(url, "api", onCloseError))
    const [api] = useState(new Api(rpc))

    return (
        <>
            <ApplicationBar api={api} rpc={rpc}/>
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
