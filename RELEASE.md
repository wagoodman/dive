# Release process


## Creating a release

**Trigger a new release with `make release`**. 

At this point you'll see a preview changelog in the terminal. If you're happy with the 
changelog, press `y` to continue, otherwise you can abort and adjust the labels on the 
PRs and issues to be included in the release and re-run the release trigger command.


## Retracting a release

If a release is found to be problematic, it can be retracted with the following steps:

- Deleting the GitHub Release
- Untag the docker images in the `docker.io` registry
- Revert the brew formula in [`joschi/homebrew-dive`](https://github.com/joschi/homebrew-dive) to point to the previous release
- Add a new `retract` entry in the go.mod for the versioned release

**Note**: do not delete release tags from the git repository since there may already be references to the release
in the go proxy, which will cause confusion when trying to reuse the tag later (the H1 hash will not match and there
will be a warning when users try to pull the new release).
