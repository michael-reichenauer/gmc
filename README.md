# gmc

Gmc is a cross-platform console UI Git client (go source), which makes it easier to use Git, especially when using a branch model similar to e.g. GitFlow. Gmc visualizes the branch structure more like one imagines the branches, instead of just showing branches exactly as the Git raw data specifies. Gmc makes it easy to toggle which branches to show and hide and simplifies using the most common git commands.

<img src="Media/gmc.gif" width="860">

## Background

Usually, Git clients visualize the repository as an overwhelming number of branches, which makes the commits history difficult to understand. As a workaround, many developers simplify git history by rebasing or squashing.

Some clients try to reduce the branching complexity by hiding commits. The Gmc approach is to make it possible to toggle which branches to show and which branches to hide. Thus a user can focus on tracking branches that really matter to them. For a developer, it might be to track just the main branch and the current working branch and for a team leader, it might be tracking main and a few selected features branches.

Gmc provides a user experience, where the visualization of branches and commits history is understandable and usable without the need for rebasing or squashing. Gmc also simplifies the usage of the most common commands by providing context menus and simplified dialogs. Gmc console window supports both key-based navigation and mouse support.

### Status

* alpha

## Download

* [Windows, Linux, Mac](https://github.com/michael-reichenauer/gmc/releases)
