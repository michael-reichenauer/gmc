import {useEffect, useState} from "react";
import {useSelector} from "react-redux";
import {IsConnected} from "../Connection/ConnectionSlice";
import {api} from "../Connection/Connection";


function commit(index, subject, author, datetime) {
    const branchIndex = 0
    const branchName = "branch" + branchIndex
    const isMerge = false

    return {index, subject, author, datetime, branchName, branchIndex, isMerge}
}

function toRepo(viewRepo) {
    const commits = viewRepo.Commits.map((c, i) => {
        const d= new Date(c.AuthorTime)
        const date = d.getFullYear() +"-" +("0"+(d.getMonth()+1)).slice(-2) + "-" + ("0" + d.getDate()).slice(-2) +
            " " + ("0" + d.getHours()).slice(-2) + ":" + ("0" + d.getMinutes()).slice(-2);
        return commit(i, c.Subject, c.Author, date, "m", 0, false);
    })
    const branches = []
    const merges = []
          return {commits, branches, merges}
}

function processChanges(changes, setRepo) {
    for (let i = 0; i < changes.length; i++) {
        console.info("change:", changes[i]);
        if (changes[i].IsStarting || changes[i].Error !== null) {
            continue
        }
        setRepo(toRepo(changes[i].ViewRepo))
    }
}

export function useRepo() {
    const isConnected = useSelector(IsConnected)
    const [callCount, setCallCount] = useState(0)
    const [repo, setRepo] = useState(null)
    const [repoID, setRepoID] = useState("")

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
                    processChanges(rsp, setRepo)
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