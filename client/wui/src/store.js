import {configureStore} from '@reduxjs/toolkit'
import {counterSlice} from "./theme";

export const store = configureStore({
    reducer: {
        //theme: themeSlice.reducer,
        counter: counterSlice.reducer,
    }
})
