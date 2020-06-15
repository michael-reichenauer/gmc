import {configureStore} from '@reduxjs/toolkit'
import {repoSlice} from "./components/Repo/RepoSlices";

export const store = configureStore({
    reducer: {
        repo: repoSlice.reducer,
    }
})
