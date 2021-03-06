import React from "react";
import Button from "@material-ui/core/Button";
import {IsConnected, IsConnecting, SetConnected, SetConnecting} from "./ConnectionSlice";
import {useDispatch, useSelector} from "react-redux";
import {useSnackbar} from "notistack";
import {RpcClient} from "../../api/RpcClient";
import {Api} from "../../api/api";
import {EventsClient} from "../../api/EventsClient";

export const rpc = new RpcClient()
export const events = new EventsClient()
export const api = new Api(rpc)


export function Connection() {
    const url = `ws://localhost:9090/api/ws`

    const dispatch = useDispatch()
    const {enqueueSnackbar, closeSnackbar} = useSnackbar();
    const onCloseError = err => {
        events.close()
        dispatch(SetConnected(false))
        enqueueSnackbar(`Connection failed`, {variant: "error", onClick: () => closeSnackbar()})
    }

    const isConnected = useSelector(IsConnected)
    const isConnecting = useSelector(IsConnecting)
    if (isConnected || isConnecting) {
        return (
            <div/>
        )
    }

    const connect = () => {
        dispatch(SetConnecting("localhost"))
        rpc.connect(url, "api", onCloseError)
            .then(() => {
                dispatch(SetConnected(true))
            })
            .catch(err => {
                dispatch(SetConnected(false))
            })
    }

    return (
        <>
            <Button
                disabled={isConnecting}
                variant="contained"
                color="primary"
                onClick={() => connect()}
            >
                Login
            </Button>
        </>
    )
}