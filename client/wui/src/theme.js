import {createMuiTheme} from "@material-ui/core/styles";
import {deepPurple, purple} from "@material-ui/core/colors";
import {createSlice} from "@reduxjs/toolkit";
import {useState} from "react";

export const useTheme = () =>{
    const [theme, setTheme]= useState(darkTheme)
    return [theme, setTheme]
}

export const darkTheme = createMuiTheme({
    palette: {
        type: "dark",
        primary: deepPurple,
        secondary: purple,
    },
});

export const lightTheme = createMuiTheme({
    palette: {
        type: "light",
        primary: deepPurple,
        secondary: purple,
    },
});

export const isLight = theme => theme===lightTheme


// export const themeSlice = createSlice({
//     name: 'theme',
//     initialState: 1,
//     reducers: {
//         setDark: state => {alert("set dark");return 0},
//         setLight: state => {alert("set light");return 1},
//     }
// })
//
// export const getTheme = state => state.theme===0 ? darkTheme:lightTheme

export const counterSlice = createSlice({
    name: 'counter',
    initialState:3,
    reducers:{
        increment: state => state+1,
        decrement: state => state-1,
    }
})


