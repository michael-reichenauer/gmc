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

    CloseRepo = () => {
        return this.rpc.Call('CloseRepo')
    }

    GetRepoChanges = () => {
        return this.rpc.Call('GetRepoChanges')
    }

    TriggerRefreshRepo = () => {
        return this.rpc.Call('TriggerRefreshRepo')
    };

    
    TriggerSearch = (text) => {
        return this.rpc.Call('TriggerSearch', text)
    }

    GetBranches = (getBranchesArgs) => {
        return this.rpc.Call('GetBranches', getBranchesArgs)
    }

    GetCommitDiff = (id) => {
        return this.rpc.Call('GetCommitDiff', id)
    }

    Commit = (message) => {
        return this.rpc.Call('Commit', message)
    }

    ShowBranch = (name) => {
        return this.rpc.Call('ShowBranch', name)
    }

    HideBranch = (name) => {
        return this.rpc.Call('HideBranch', name)
    }

    Checkout = (checkoutArgs) => {
        return this.rpc.Call('Checkout', checkoutArgs)
    }

    PushBranch = (name) => {
        return this.rpc.Call('PushBranch', name)
    }

    PullCurrentBranch = () => {
        return this.rpc.Call('PullCurrentBranch')
    }

    MergeBranch = (name) => {
        return this.rpc.Call('MergeBranch', name)
    }

    CreateBranch = (name) => {
        return this.rpc.Call('CreateBranch', name)
    }

    DeleteBranch = (name) => {
        return this.rpc.Call('DeleteBranch', name)
    }


    Catch = promise => {
        promise.then(() => {
        })
            .catch(err => {
                console.warn("Failed:", err)
            })
    }
}
