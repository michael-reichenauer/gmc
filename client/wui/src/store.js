import {configureStore} from '@reduxjs/toolkit'
import {repoSlice} from "./components/Repo/RepoSlices";
import {connectionSlice} from "./components/Connection/ConnectionSlice";

export const store = configureStore({
    reducer: {
        repo: repoSlice.reducer,
        connection: connectionSlice.reducer
    }
})
