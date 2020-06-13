import {createMuiTheme} from "@material-ui/core/styles";
import red from "@material-ui/core/colors/red";
import blue from "@material-ui/core/colors/blue";
import {deepPurple, green, purple} from "@material-ui/core/colors";


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

