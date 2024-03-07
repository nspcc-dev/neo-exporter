# Release instructions

## Pre-release checks

These should run successfully:
* `make bin`;
* `make image`;

## Release steps

1. Update `CHANGELOG.md` and `VERSION` files, create PR.

2. After successful PR merge:
   * Create Github release, copy changelog there.
   * Wait for autobuild to complete successfully.
