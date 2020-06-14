import React from "react";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
//import Switch from "@material-ui/core/Switch";
import SearchIcon from "@material-ui/icons/Search";
import InputBase from "@material-ui/core/InputBase";
import IconButton from "@material-ui/core/IconButton";
import MenuIcon from "@material-ui/icons/Menu";
import AppBar from "@material-ui/core/AppBar";
import {useAppBarStyles} from "../appStyles";
import {useDispatch, useSelector} from "react-redux";
//import {getTheme, lightTheme, themeSlice} from "../theme";
//import {AppContext} from "../appContext";
//import { useTheme} from "../theme";
import Switch from "@material-ui/core/Switch";
import {counterSlice} from "../theme";

export const ApplicationBar = props => {
    const classes = useAppBarStyles();
    const counter = useSelector(state => state.counter)
    const dispatch = useDispatch()
    return (
        <AppBar position="static">
            <Toolbar>
                <Typography className={classes.title} variant="h6" noWrap>
                    gmc {counter}
                </Typography>
                <Switch
                    checked={counter===0}
                    onChange={() => dispatch(counterSlice.actions.increment())}
                />

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