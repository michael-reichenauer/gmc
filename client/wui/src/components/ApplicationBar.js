import React from "react";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import Switch from "@material-ui/core/Switch";
import SearchIcon from "@material-ui/icons/Search";
import InputBase from "@material-ui/core/InputBase";
import IconButton from "@material-ui/core/IconButton";
import MenuIcon from "@material-ui/icons/Menu";
import AppBar from "@material-ui/core/AppBar";
import {useAppBarStyles} from "../appStyles";
import {useDispatch, useSelector} from "react-redux";
import {connectionSlice, IsConnected, IsConnecting} from "./Connection/ConnectionSlice";


export const ApplicationBar = props => {
    const {api, rpc} = props
    const classes = useAppBarStyles();

    const dispatch = useDispatch()
    const isConnected = useSelector(IsConnected)
    const isConnecting = useSelector(IsConnecting)


    const connect = () => {
        dispatch(connectionSlice.actions.setConnecting("localhost"))
        rpc.Connect()
            .then(() => {
                dispatch(connectionSlice.actions.setConnected(true))
            })
            .catch(err => {
                dispatch(connectionSlice.actions.setError("Failed to connect"))
            })
    }

    const close = () => {
        dispatch(connectionSlice.actions.setError(""))
        rpc.Close()
        dispatch(connectionSlice.actions.setError(""))
    }

    const toggle = () => {
        if (isConnected) {
            close()
        } else {
            connect()
        }
    }

    return (
        <AppBar position="static">
            <Toolbar>
                <Typography className={classes.title} variant="h6" noWrap>
                    gmc
                </Typography>
                <Switch disabled={isConnecting} checked={isConnected} onChange={() => toggle()}/>
                <div className={classes.search}>
                    <div className={classes.searchIcon}>
                        <SearchIcon/>
                    </div>
                    <InputBase
                        placeholder="Search…"
                        classes={{
                            root: classes.inputRoot,
                            input: classes.inputInput,
                        }}
                        inputProps={{'aria-label': 'search'}}
                    />
                </div>

                <IconButton
                    edge="start"
                    className={classes.menuButton}
                    color="inherit"
                    aria-label="open drawer"
                >
                    <MenuIcon/>
                </IconButton>
            </Toolbar>
        </AppBar>
    )
}