# maildir_tools

### Dumper
The CLI tool _dumper_ is running `du -s` in an endless loop, dumping the current size of a directory in bytes. Once it gets a system call all dump files are zipped and uploaded to a GCS bucket. These files in GCS can be compared across machines to compute the replication lag.

### Visualizer 

The CLI tool _visualizer_ takes two zipped files, unzips them in memory and builds a matplotlib based python file to compare the replication lag visually.
