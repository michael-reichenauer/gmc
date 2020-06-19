const JsonRPC = require('simple-jsonrpc-js');
const noArg = [0]

export class RpcClient {
    constructor(url, methodPrefix) {
        this.url = url
        this.methodPrefix = methodPrefix + "."
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
                    console.info(`Connection close was clean for ${this.url}`);
                } else {
                    console.warn(`Connection unexpected closed for ${this.url}`);
                }
                console.info(`close code : ${event.code}, reason: ${event.reason}`);
            };
        });
    };

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


