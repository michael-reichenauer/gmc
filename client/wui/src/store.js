import {configureStore} from '@reduxjs/toolkit'
import {connectionSlice} from "./components/Connection/ConnectionSlice";

export const store = configureStore({
    reducer: {
        connection: connectionSlice.reducer
    }
})
