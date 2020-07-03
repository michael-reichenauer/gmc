export class EventsClient {
    constructor() {
        console.info("Creating events client")
    }

    connect = (url, onEvent) => {
        return new Promise((resolve, reject) => {
            const eventSource = new EventSource(url)

            eventSource.onopen = e => {
                console.warn("onopen:", e)
                this.eventSource = eventSource
                this.url =url
                resolve()
            }

            eventSource.onmessage = function (msg) {
                console.warn("event:", msg.data)
                const event = JSON.parse(msg.data);
                onEvent(event)
            }

            eventSource.onerror = error => {
                console.warn(`Connect to ${url} failed, ${error}`);
                reject(error)
            }
        });
    };

    close = ()=>{
        console.info(`Closing connection to ${this.url} ...`);
        if (this.eventSource && this.eventSource.readyState === EventSource.OPEN) {
            this.eventSource.close()
        }
    }
}