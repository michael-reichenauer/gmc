import {createSlice} from "@reduxjs/toolkit";
import {mockRepo} from "./mockData";

export const repoSlice = createSlice({
    name: 'repo',
    initialState: {none:true},
    reducers:{
        set: (state, action) => action.payload,
    }
})
