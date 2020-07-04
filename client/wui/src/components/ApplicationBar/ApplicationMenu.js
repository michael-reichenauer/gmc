import React, {useState} from "react";
import {rpc} from "../Connection/Connection";
import {IsConnected, SetConnected} from "../Connection/ConnectionSlice";
import {useDispatch, useSelector} from "react-redux";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import IconButton from "@material-ui/core/IconButton";
import MenuIcon from "@material-ui/icons/Menu";
import makeStyles from "@material-ui/core/styles/makeStyles";


const useMenuStyles = makeStyles((theme) => ({
    menuButton: {
        marginRight: theme.spacing(2),
    },
}));


export function ApplicationMenu() {
    const classes = useMenuStyles();
    const dispatch = useDispatch()
    const isConnected = useSelector(IsConnected)

    const [menu, setMenu] = useState(null);

    const handleLogout = () => {
        setMenu(null);
        rpc.close()
        dispatch(SetConnected(false))
    };

    return (
        <>
            <IconButton
                edge="start"
                className={classes.menuButton}
                color="inherit"
                onClick={e => setMenu(e.currentTarget)}
            >
                <MenuIcon/>
            </IconButton>
            <Menu
                anchorEl={menu}
                keepMounted
                open={Boolean(menu)}
                onClose={() => setMenu(null)}
            >
                <MenuItem disabled={!isConnected} onClick={handleLogout}>Logout</MenuItem>
            </Menu>
        </>
    )
}