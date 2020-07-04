import {createMuiTheme} from "@material-ui/core/styles";
import {deepPurple, purple} from "@material-ui/core/colors";


export const darkTheme = createMuiTheme({
    palette: {
        type: "dark",
        primary: deepPurple,
        secondary: purple,
      
    },
});

export const lightTheme = createMuiTheme({
    palette: {
        type: "light",
        primary: deepPurple,
        secondary: purple,
    },
});

export const isLight = theme => theme===lightTheme
