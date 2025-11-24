# Requirements

- accepts a path which must be a directory, and operates only on items contained in that directory (does not recursively descend through other directories)
- specify a number of files to keep N and keep only the newest N files
  - if this is specified in combination with other options, treat this as the minimum number of files to keep and keep the N newest regardless of whether or not they would otherwise be removed
- specify a unix timestamp and remove any files older than said timestamp
- specify an interval (seconds, minutes, hours, days) and remove any files older than that interval
- allow a dry-run which only prints files to be removed (may be useful in actual usage to pipe to other programs)
- switch to operate on directories (tool should probably only operate on directories or files in any single invocation, not both)
- by default operate based on modified date
- switch to operate based on creation date
- switch to move items to a destination rather than delete them
- confirm by default, switch to execute without confirmation
- experiment with how much output is reasonable, and determine whether just a verbose flag is necessary or more granular controls
- should only operate on actual files in the given path, if those items are symlinks then delete only the symlink
  - perhaps include an option to consider the created/modified time on the symlink target when deciding what to do with the symlink itself
