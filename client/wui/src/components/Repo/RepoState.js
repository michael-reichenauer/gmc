import {useEffect, useState} from "react";
import {useSelector} from "react-redux";
import {IsConnected} from "../Connection/ConnectionSlice";
import {api} from "../Connection/Connection";
import {mockRepo} from "./mockData";



export function useRepo() {
    const isConnected = useSelector(IsConnected)
    const [callCount, setCallCount] = useState(0)
    const [repo, setRepo] = useState(null)

    useEffect(() => {
        if (!isConnected) {
            console.info("useRepo: Not connected")
            if (callCount !== 0) {
                setCallCount(0)
            }
            return
        }
        console.info("useRepo: Connected")
        console.info("useRepo: Get changes ...")
        api.GetRepoChanges(["2"])
            .then(rsp => {
                    console.info("useRepo: Got changes", rsp);
                    //commits= rsp[0].viewport.
                    setCallCount(callCount + 1)
                    setRepo(mockRepo)
                }
            )
            .catch(err => {
                console.warn("useRepo: Error", err)
            })
        if (callCount === 0) {
            console.info("useRepo: Trigger refresh")
            api.TriggerRefreshRepo()
        }
    }, [isConnected, callCount])

    return repo
}