import {createSlice} from "@reduxjs/toolkit";


const initialState = {
    isConnecting : false,
    isConnected: false,
    url:"",
}

export const connectionSlice = createSlice({
    name: 'connection',
    initialState:initialState,
    reducers:{
        setConnecting: (state, action) => {return {isConnected: false, isConnecting: true, url: action.payload}},
        setConnected: (state, action) => {return {isConnected: true, isConnecting: false, url: state.url}},
        setDisconnected: (state, action) => {return {isConnected: false, isConnecting: false, url:""}},
    }
})

export const IsConnecting = state =>state.connection.isConnecting
export const IsConnected = state =>state.connection.isConnected
