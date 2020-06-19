import {configureStore} from '@reduxjs/toolkit'
import {repoSlice} from "./components/Repo/RepoSlices";
import {connectingSlice} from "./components/Connecting/connectingSlice";

export const store = configureStore({
    reducer: {
        repo: repoSlice.reducer,
        connecting: connectingSlice.reducer
    }
})
