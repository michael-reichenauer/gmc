import {createMuiTheme} from "@material-ui/core/styles";
import red from "@material-ui/core/colors/red";
import blue from "@material-ui/core/colors/blue";
import {green, purple} from "@material-ui/core/colors";


const theme = createMuiTheme({
    palette: {
       type:"dark",
        primary: purple
    },
});

export default theme