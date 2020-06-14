import {configureStore} from '@reduxjs/toolkit'
import {counterSlice} from "./theme";
import {repoSlice} from "./components/Repo/RepoSlices";

export const store = configureStore({
    reducer: {
        //theme: themeSlice.reducer,
        counter: counterSlice.reducer,
        repo: repoSlice.reducer,
    }
})
