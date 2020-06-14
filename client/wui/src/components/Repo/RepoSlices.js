import {createSlice} from "@reduxjs/toolkit";
import {mockRepo} from "./mockData";

export const repoSlice = createSlice({
    name: 'repo',
    initialState: mockRepo,
    reducers:{
        set: (state, action) => action.payload,
    }
})
