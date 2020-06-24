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
        setConnected: (state, action) => {
            const isConnected = action.payload
            const url = isConnected ? state.url:""
            return {isConnected: isConnected, isConnecting: false, url: url}},
    }
})

export const SetConnected = isConnected=> connectionSlice.actions.setConnected(isConnected)
export const SetConnecting = url => connectionSlice.actions.setConnecting(url)

export const IsConnecting = state =>state.connection.isConnecting
export const IsConnected = state =>state.connection.isConnected
