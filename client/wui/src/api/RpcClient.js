const JsonRPC = require('simple-jsonrpc-js');
const noArg = [0]

export class RpcClient {
    constructor() {
        console.info("Creating rpc client")
    }

    connect = (url, eventsUrl, methodPrefix, onCloseError) => {
        this.url = url
        this.methodPrefix = methodPrefix + "."
        this.onCloseError = onCloseError
        this.jsonRPC = null

        return new Promise((resolve, reject) => {
            const jsonRPC = new JsonRPC();

            console.info(`Connecting to ${url} ...`);
            this.socket = new WebSocket(url);

            this.socket.onopen = () => {
                console.info(`Connected to ${url}`);
                this.jsonRPC = jsonRPC
                this.eventClient = new EventSource(eventsUrl)
                this.eventClient.onmessage = function (msg) {
                    console.log("event:", msg.data)
                }
                this.eventClient.onerror = error =>{
                    console.warn(`Connect to ${eventsUrl} failed, ${error}`);
                }
                this.eventClient.onopen = e => {
                    console.warn("onopen:", e)
                }
                resolve()
            };

            this.socket.onerror = error => {
                console.warn(`Connect to ${url} failed, ${error}`);
                reject(error)
            };


            this.socket.onmessage = event => {
                jsonRPC.messageHandler(event.data);
            };

            jsonRPC.toStream = (msg) => {
                this.socket.send(msg);
            };

            this.socket.onclose = event => {
                if (this.eventClient && this.eventClient.readyState !== EventSource.CLOSED){
                    this.eventClient.close()
                }
                if (event.wasClean) {
                    console.info(`Closed connection to ${url}`);
                } else {
                    const error = `code : ${event.code}, reason: ${event.reason}`

                    if (this.jsonRPC == null) {
                        console.warn(`Failed to connect to ${url}, ${error}`)
                    } else {
                        console.warn(`Connection unexpected closed for ${url}, ${error}`);
                    }
                    if (this.onCloseError) {
                        this.onCloseError(error)
                    }
                }
            };
        });
    };

    close = () => {
        console.info(`Closing connection to ${this.url} ...`);
        if (this.eventClient && this.eventClient.readyState === EventSource.OPEN) {
            this.eventClient.close()
        }
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.close()
        }
        this.url = ""
        this.methodPrefix = ""
    }

    call = (method, param) => {
        method = this.methodPrefix + method
        if (param === undefined) {
            param = noArg
            console.info(`Calling: ${method} ()`)
        } else {
            console.info(`Calling: ${method} (`, param, ")")
            param = [param]
        }

        return new Promise((resolve, reject) => {
            this.jsonRPC.call(method, param)
                .then(rsp => {
                    console.info("OK:", method)
                    resolve(rsp)
                })
                .catch(err => {
                    console.warn("Failed:", method, err)
                    reject(err)
                })
        })
    };

    static isEmpty = (value) => {
        if (RpcClient.isObject(value)) {
            for (let idx in value) {
                if (value.hasOwnProperty(idx)) {
                    return false;
                }
            }
            return true;
        }
        if (Array.isArray(value)) {
            return !value.length;
        }
        return !value;
    };

    static isObject = (value) => {
        const type = typeof value;
        return value != null && (type === 'object' || type === 'function');
    };
}


