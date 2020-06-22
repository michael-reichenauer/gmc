const JsonRPC = require('simple-jsonrpc-js');
const noArg = [0]

export class RpcClient {
    constructor(url, methodPrefix, onCloseError) {
        this.url = url
        this.methodPrefix = methodPrefix + "."
        this.onCloseError = onCloseError
    }

    Connect = () => {
        return new Promise((resolve, reject) => {
            const jsonRPC = new JsonRPC();

            console.info(`Connecting to ${this.url} ...`);
            this.socket = new WebSocket(this.url);

            this.socket.onopen = () => {
                console.info(`Connected to ${this.url}`);
                this.jsonRPC = jsonRPC
                resolve()
            };

            this.socket.onerror = error => {
                console.warn(`Connection error for ${this.url}, ${error.message}`);
                reject(error)
            };


            this.socket.onmessage = event => {
                jsonRPC.messageHandler(event.data);
            };

            jsonRPC.toStream = (msg) => {
                this.socket.send(msg);
            };

            this.socket.onclose = event => {
                if (event.wasClean) {
                    console.info(`Closed connection to ${this.url}`);
                } else {
                    const error = `code : ${event.code}, reason: ${event.reason}`
                    console.warn(`Connection unexpected closed for ${this.url}, ${error}`);
                    if (this.onCloseError){
                        this.onCloseError(error)
                    }
                }
            };
        });
    };

    Close = ()=>{
        console.info(`Close connection to ${this.url}`);
        this.socket.close()
    }

    Call = (method, params) => {
        method = this.methodPrefix + method
        console.info("Calling:", method, params)
        if (RpcClient.isEmpty(params)) {
            // Make sure there is at least one noArg parameter as required by json rps server
            params = noArg
        }

        return new Promise((resolve, reject) => {
            this.jsonRPC.call(method, params)
                .then(rsp => {
                    console.info("OK:", method, rsp)
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


