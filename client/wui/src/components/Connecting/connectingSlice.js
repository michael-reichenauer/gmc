import {createSlice} from "@reduxjs/toolkit";
import {mockRepo} from "../Repo/mockData";

export const connectingSlice = createSlice({
    name: 'connecting',
    initialState: "",
    reducers:{
        set: (state, action) => action.payload,
    }
})
