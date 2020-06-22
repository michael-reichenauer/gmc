import {useState} from "react";


export function useRepo() {
    const [repo, setRepo] = useState(null)

    return [repo, setRepo]
}