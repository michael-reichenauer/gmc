

export class Api {
    constructor(url, cb) {
        this.url = url
        this.cb = cb
    }

    connect = () => {
        console.log("connecting ...");
        this.socket = new WebSocket(this.url);

        this.socket.onopen = () => {
            console.log("Successfully Connected");
        };

        this.socket.onmessage = msg => {
            console.log("onmessage", msg);
            this.cb(msg);
        };

        this.socket.onclose = event => {
            console.log("Socket Closed Connection: ", event);
        };

        this.socket.onerror = error => {
            console.log("Socket Error: ", error);
        };
    };

    sendMsg = msg => {
        console.log("sending msg: ", msg);
        this.socket.send(msg);
    };
}

