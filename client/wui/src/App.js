// App.js
import React, {useState} from "react";

import Toolbar from "@material-ui/core/Toolbar";
import IconButton from "@material-ui/core/IconButton";
import MenuIcon from '@material-ui/icons/Menu';
import AppBar from "@material-ui/core/AppBar";
import SearchIcon from '@material-ui/icons/Search';
import InputBase from "@material-ui/core/InputBase";
import makeStyles from "@material-ui/core/styles/makeStyles";
import Paper from "@material-ui/core/Paper";

import Switch from "@material-ui/core/Switch";
import TableContainer from "@material-ui/core/TableContainer";
import ThemeProvider from "@material-ui/styles/ThemeProvider";
import createMuiTheme from "@material-ui/core/styles/createMuiTheme";
import Typography from "@material-ui/core/Typography";
import {fade} from "@material-ui/core";
import Table from "@material-ui/core/Table";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";
import TableBody from "@material-ui/core/TableBody";


const useStyles = makeStyles((theme) => ({
    table: {
        minWidth: 650,
    },
    root: {
        flexGrow: 1,
    },
    menuButton: {
        marginRight: theme.spacing(2),
    },
    title: {
        flexGrow: 1,
        display: 'none',
        [theme.breakpoints.up('sm')]: {
            display: 'block',
        },
    },
    text: {
        fontFamily: "monospace",

    },
    search: {
        position: 'relative',
        borderRadius: theme.shape.borderRadius,
        backgroundColor: fade(theme.palette.common.white, 0.15),
        '&:hover': {
            backgroundColor: fade(theme.palette.common.white, 0.25),
        },
        marginLeft: 0,
        width: '100%',
        [theme.breakpoints.up('sm')]: {
            marginLeft: theme.spacing(1),
            width: 'auto',
        },
    },
    searchIcon: {
        padding: theme.spacing(0, 2),
        height: '100%',
        position: 'absolute',
        pointerEvents: 'none',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
    },
    inputRoot: {
        color: 'inherit',
    },
    inputInput: {
        padding: theme.spacing(1, 1, 1, 0),
        // vertical padding + font size from searchIcon
        paddingLeft: `calc(1em + ${theme.spacing(4)}px)`,
        transition: theme.transitions.create('width'),
        width: '100%',
        [theme.breakpoints.up('sm')]: {
            width: '12ch',
            '&:focus': {
                width: '20ch',
            },
        },
    },
}));

function createData(name, calories, fat, carbs, protein) {
    return {name, calories, fat, carbs, protein}
}

const rows = [
    createData('Frozen', 159, 6.0, 24, 4.0),
    createData('asdf', 19, 6.0, 24, 4.0),
    createData('csdff', 23, 6.0, 24, 4.0),
    createData('Froxzvzen', 15, 6.0, 24, 4.0),
    createData('Frozen', 3, 6.0, 24, 3.0),
]

const App = () => {
    const [lightMode, setLightMode] = useState(false)
    const darkTheme = createMuiTheme({
        palette: {
            type: "dark",
        },
    });
    const lightTheme = createMuiTheme({
        palette: {},
    });

    const classes = useStyles();
    let text = '┠┬      Merge branch \'branches/newFeat\' into develop (1)\n                    ┃└──┰ * Some more cleaning (2)\n                    ┠┬  ┃   Merge branch \'branches/diff\' into develop (1)\n                    ┃└┰ ┃   Adjust commitVM diff (3)\n                    ┃┌┸ ┃   Update git to 2.23 (3)\n                    ┃│  ┠   Fixing a bug (2)\n                    ┃│ ┌┸   Clean code (2)\n                    █┴─┘    Merge branch \'branches/branchcommit\' into develop (1)\n                    ┠       fix tag names with strange char ending (1)\n                    ┠       Clean build script (1)\n                    ┠┐      Merge branch \'branches/NewBuild\' into develop (1)\n                    ┃└┰     Update some cake tools (4)\n                    ┃┌┸     Add Cake build script support (4)\n                    ┠┼      Merge branch \'branches/FixIssues\' into develop (1)\n                    ┃└┰     Adjust file monitor logging (5)\n                    ┃ ┠     Adjust expected git version (5)\n                    ┃ ┠     Use git 2.20.0  (5)\n                    ┃ ┠     Clean diff temp files (5)\n                    ┃┌┺     Fix missing underscore char in details file list (5)\n                    ┠┴      Version 0.144 (1)\n                    ┠       Some text  (1)'
    let newText = text.split('\n').map((item, i) => {
        return <p key={i}>{item}</p>;
    });

    return (
        <ThemeProvider theme={lightMode ? lightTheme : darkTheme}>
            <Paper style={{height: "100vh"}}>
                <AppBar position="static">
                    <Toolbar>
                        <Typography align={"left"} className={classes.title} variant="h6" noWrap>
                            gmc
                        </Typography>
                        <Switch checked={lightMode} onChange={() => setLightMode(!lightMode)}/>
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
                <TableContainer component={Paper}>
                    <Table className={classes.table} size="small">
                        <TableHead>
                            <TableRow>
                                <TableCell>Dessert</TableCell>
                                <TableCell align="right">Calories</TableCell>
                                <TableCell align="right">Fat</TableCell>
                                <TableCell align="right">Carbs</TableCell>
                                <TableCell align="right">Protein</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {rows.map((row) => (
                                <TableRow key={row.name}>
                                    <TableCell component="th" scope="row">{row.name}</TableCell>
                                    <TableCell align={"right"}>{row.calories}</TableCell>
                                    <TableCell align={"right"}>{row.fat}</TableCell>
                                    <TableCell align={"right"}>{row.carbs}</TableCell>
                                    <TableCell align={"right"}>{row.protein}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>

                </TableContainer>
            </Paper>
        </ThemeProvider>
    );
}

export default App;