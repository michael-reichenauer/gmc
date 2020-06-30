import {useEffect, useState} from "react";
import {useSelector} from "react-redux";
import {IsConnected} from "../Connection/ConnectionSlice";
import {api} from "../Connection/Connection";
import {mockRepo} from "./mockData";


export function useRepo() {
    const isConnected = useSelector(IsConnected)
    const [callCount, setCallCount] = useState(0)
    const [repo, setRepo] = useState(null)
    const [repoID, setRepoID] = useState("")

    function processChanges(changes) {
        for (let i = 0; i < changes.length; i++) {
            console.info("change:", changes[i]);
            if (changes[i].IsStarting || changes[i].Error !== null) {
                continue
            }
            setRepo(mockRepo)
        }
    }

    useEffect(() => {
        if (!isConnected) {
            console.info("useRepo: Not connected")
            if (callCount !== 0) {
                // Reset repo data after disconnect
                console.info("useRepo: rest call count");
                setCallCount(0)
                setRepoID("")
            }
            return
        }

        if (repoID === "") {
            api.OpenRepo("")
                .then(rsp => setRepoID(rsp))
                .catch(err => console.warn("Failed to open repo", err))
            return
        }

        api.GetRepoChanges(repoID)
            .then(rsp => {
                    console.info("useRepo: Got changes", rsp);
                    //commits= rsp[0].viewport.
                    processChanges(rsp)
                    setCallCount(callCount + 1)
                }
            )
            .catch(err => {
                console.warn("useRepo: Error", err)
            })

        if (callCount === 0) {
            // First time after connect
            api.TriggerRefreshRepo(repoID)
        }
    }, [isConnected, callCount, repoID])

    return repo
}