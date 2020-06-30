export class Api {
    constructor(rpcClient) {
        this.rpc = rpcClient
    }

    GetRecentWorkingDirs = () => {
        return this.rpc.Call('GetRecentWorkingDirs')
    }

    GetSubDirs = (dirPath) => {
        return this.rpc.Call('GetSubDirs', dirPath)
    }

    OpenRepo = (path) => {
        return this.rpc.Call('OpenRepo', path)
    }

    CloseRepo = (repoID) => {
        return this.rpc.Call('CloseRepo', repoID)
    }

    GetRepoChanges = (repoID) => {
        return this.rpc.Call('GetRepoChanges', repoID)
    }

    TriggerRefreshRepo = (repoID) => {
        this.Catch(this.rpc.Call('TriggerRefreshRepo', repoID))
    };

    TriggerSearch = (search) => {
        return this.rpc.Call('TriggerSearch', search)
    }

    GetBranches = (getBranches) => {
        return this.rpc.Call('GetBranches', getBranches)
    }

    GetCommitDiff = (commitDiffInfo) => {
        return this.rpc.Call('GetCommitDiff', commitDiffInfo)
    }

    Commit = (commitInfo) => {
        return this.rpc.Call('Commit', commitInfo)
    }

    ShowBranch = (branchName) => {
        return this.rpc.Call('ShowBranch', branchName)
    }

    HideBranch = (branchName) => {
        return this.rpc.Call('HideBranch', branchName)
    }

    Checkout = (checkout) => {
        return this.rpc.Call('Checkout', checkout)
    }

    PushBranch = (branchName) => {
        return this.rpc.Call('PushBranch', branchName)
    }

    PullCurrentBranch = (repoID) => {
        return this.rpc.Call('PullCurrentBranch', repoID)
    }

    MergeBranch = (branchName) => {
        return this.rpc.Call('MergeBranch', branchName)
    }

    CreateBranch = (branchName) => {
        return this.rpc.Call('CreateBranch', branchName)
    }

    DeleteBranch = (branchName) => {
        return this.rpc.Call('DeleteBranch', branchName)
    }


    Catch = promise => {
        promise
            .then(() => {

            })
            .catch(err => {
                console.warn("Failed xxx:", err)
            })
    }
}
