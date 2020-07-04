import {useEffect, useState} from "react";
import {useSelector} from "react-redux";
import {IsConnected} from "../Connection/ConnectionSlice";
import {api, events} from "../Connection/Connection";


function commit(index, subject, author, datetime) {
    const branchIndex = 0
    const branchName = "branch" + branchIndex
    const isMerge = false

    return {index, subject, author, datetime, branchName, branchIndex, isMerge}
}

function toRepo(viewRepo) {
    const commits = viewRepo.Commits.map((c, i) => {
        const d = new Date(c.AuthorTime)
        const date = d.getFullYear() + "-" + ("0" + (d.getMonth() + 1)).slice(-2) + "-" + ("0" + d.getDate()).slice(-2) +
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
        console.info("New repo")
        setRepo(toRepo(changes[i].ViewRepo))
    }
}

function getRepoChanges(repoID, setRepo) {
    console.info(`getRepoChanges:Getting repo changes from ${repoID} ...`)
    const onEvents = event =>{
        console.warn("On event: ", event)
    }

    events.connect(`http://localhost:9090/api/events/${repoID}`, onEvents)
        .then(()=> api.TriggerRefreshRepo(repoID))
        .catch(err=>{console.warn("Failed to connect")})
    

    // api.GetRepoChanges(repoID)
    //     .then(rsp => {
    //             console.info("getRepoChanges: Got changes", rsp);
    //             //commits= rsp[0].viewport.
    //             processChanges(rsp, setRepo)
    //             //setCallCount(callCount + 1)
    //             getRepoChanges(repoID, setRepo)
    //         }
    //     )
    //     .catch(err => {
    //         console.warn("getRepoChanges: Error", err)
    //     })
}

export function useRepo() {
    const isConnected = useSelector(IsConnected)
    //const [callCount, setCallCount] = useState(0)
    const [repo, setRepo] = useState(null)
    const [repoID, setRepoID] = useState("")

    useEffect(() => {
        if (!isConnected) {
            console.info("useRepo: Not connected")
            if (repoID !== "") {
                // Reset repo data after disconnect
                console.info("useRepo: rest call count");
                setRepoID("")
            }
            return
        }

        if (repoID === "") {
            api.OpenRepo("")
                .then(repoID => {
                    setRepoID(repoID)
                    getRepoChanges(repoID, setRepo)
                })
                .catch(err => console.warn("Failed to open repo", err))
            return
        }



    }, [isConnected, repoID])

    return repo
}