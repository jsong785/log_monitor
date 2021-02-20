Observability of log files

#TODO
## implement faster read or truncate safe reads
## implement http service tests
## implement http load tests

# Read a line in reverse (perhaps optimize)
# Read N lines in reverse (perhaps optimize)

# Read n lines from a file in reverse (and benchmark)
# Performance load testing, Read n lines from a file in reverse
## benchmark greedy reads

# POC; should pass based on what is written in "The Linux Programming Interface"
## Read n lines from file; during the middle the file is deleted (remove() and unlink()); should still pass.
## Read n lines from file; during the middle the file is moved (renamed); should still pass.
## Read n lines from file; during the middle the file is truncated to 0 (this may cause it to fail)

# Create function that sets up http server with given injections and gives the appropriate rest responses
## Ensure that responses align with REST principles; which may include self-discovery api "links"; and not allowing browser to cache responses.
# Tie this http server with real injections from the above implementations; test.

Hypothesis:
The OS could provide a lot of the optimizations through caching reads and scheduling them in an optimal matter. Only optimize if this isn't enough.

A file descriptor open on a file will provide guarantees in the event of a file rotation (file move); or perhaps some other process attempting to delete it (remove() or unlink(). Currently unsure of a rotation strategy involving truncate to zero (will investigate). Currently unsure of of syslog behavior if it fails to rotate a file due to a file descriptor being open on it as well (will investigate).

Expanding upon the REST api to include features such as pagination may not be necessary if appropriate measures are taken; such as regular file rotation; which may prevent a file from getting too large.

If file descriptor availibility becomes an issue, create a file descriptor pool.

If file truncation to zero during a read becomes an issue; determine if a "as read" response is appropriate; or simply copy the file to a unique place on the disc and read from there. This assumes no rogue process edits the files in this special area.

If files do indeed get too big, consider pagination. This could be a custom sha token generated for the user. The file can be copied with the sha appended to the filename (take care in the event of really long filename?). The next REST api request to this file with a token and page can start where it leaves off. This assumes no rogue process will interfere with copied files in this special area.

I do not think caching file contents will be helpful, Linux already does this for you.

