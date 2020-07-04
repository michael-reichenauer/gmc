export class Api {
    constructor(rpcClient) {
        this.rpc = rpcClient
    }

    GetRecentWorkingDirs = () => {
        return this.rpc.call('GetRecentWorkingDirs')
    }

    GetSubDirs = (dirPath) => {
        return this.rpc.call('GetSubDirs', dirPath)
    }

    OpenRepo = (path) => {
        return this.rpc.call('OpenRepo', path)
    }

    CloseRepo = (repoID) => {
        return this.rpc.call('CloseRepo', repoID)
    }

    GetRepoChanges = (repoID) => {
        return this.rpc.call('GetRepoChanges', repoID)
    }

    TriggerRefreshRepo = (repoID) => {
        this.catch(this.rpc.call('TriggerRefreshRepo', repoID))
    };

    TriggerSearch = (search) => {
        return this.rpc.call('TriggerSearch', search)
    }

    GetBranches = (getBranches) => {
        return this.rpc.call('GetBranches', getBranches)
    }

    GetCommitDiff = (commitDiffInfo) => {
        return this.rpc.call('GetCommitDiff', commitDiffInfo)
    }

    Commit = (commitInfo) => {
        return this.rpc.call('Commit', commitInfo)
    }

    ShowBranch = (branchName) => {
        return this.rpc.call('ShowBranch', branchName)
    }

    HideBranch = (branchName) => {
        return this.rpc.call('HideBranch', branchName)
    }

    Checkout = (checkout) => {
        return this.rpc.call('Checkout', checkout)
    }

    PushBranch = (branchName) => {
        return this.rpc.call('PushBranch', branchName)
    }

    PullCurrentBranch = (repoID) => {
        return this.rpc.call('PullCurrentBranch', repoID)
    }

    MergeBranch = (branchName) => {
        return this.rpc.call('MergeBranch', branchName)
    }

    CreateBranch = (branchName) => {
        return this.rpc.call('CreateBranch', branchName)
    }

    DeleteBranch = (branchName) => {
        return this.rpc.call('DeleteBranch', branchName)
    }


    catch = promise => {
        promise
            .then(() => {

            })
            .catch(err => {
                console.warn("Failed xxx:", err)
            })
    }
}
