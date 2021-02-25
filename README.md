# Observability of log files

## building
cd to main and `go build`

## running the server
To run, package main has the http server.
Without any arguments supplied, it runs on "localhost:8080" on "/var/log" withe writing request timeout of 2000 milliseconds. Reading a client request timeout is hardcoded to 100ms. 

The supported arguments are:
- dir="some_dir": directory to watch, trailing slash does not matter.
- timeout=NUM: write request timeout in milliseconds.
- addr="": [address:port] to run on.

It is assumed that the user knows which files to query for.
The supported query commands are:
- lines
- filter

Ex: 
- http://localhost:8080/file?lines=100
- http://localhost:8080/file?filter=abc

Combinations are supported as well; in any order; but the underlying code will do lines first, then filiter
Ex:
- http://localhost:8080/file?lines=100&filter=abc

## design
I spent most of the time attempting to optimize the file reading capabilities of the system.
I am getting worse performance than `tail -n 100000 large_file | tac` on my home computer, but on a high powered workstation, I am exceeded performance of the above.

On the high powered workstation Intel Core-i7 2.6ghz with 32gigs of ram; and SSD hard drive. (note: this happens to be a windows machine and I'm running this benchmark under WSL (Windows Subsystem for Linux))

- Reading 100K lines `time tail -n 100000 LARGE_FILE | tac > /dev/null`; averages 36ms.
- cd to `file_reader` and run `go test -count=1 -bench=BenchmarkLargeFile_SingleRequestChunk100K`; averages 9.4ms!
  - `go test -count=1 -bench=BenchmarkLargeFile_ManyRequestsChunk` which is the equivalent of above, but 1000 concurrent requests; 40129ms per run; or 40ms per run.
- diffing the output between the linux command, and the two versions of the reverse n lines reader written (one is slower), resolve to be no difference.

## file reading
### synchronous issues
There are 5 different types of problems that may occur when reading log files. They are:
- reading from a log file as writing occurs to the end of it; reading may occur in the middle of a new line being written.
- reading from a log file as a file is deleted (remove or unlink).
- reading from a log file as a file is being renamed.
- reading from a log file as a file is being moved.
- reading from a file as it is being truncated in place (may happen if in place truncation scheme is used for file rotation in syslog).

Here is how they are handled:
- POSIX defines a line as ending with '\n'; vim will automatically insert a new line when saving a file with only "one line written" (no return). Various linux programs would break if this wasn't the case. In reverse, every line must start with '\n'; otherwise it isn't a line, or reading is done concurrently with a write. Just simply go back to the first '\n'.
- Remove/unlink/rename/moving a file is handled by the linux programming interface. The file will still be around if a file descriptor is open on it. In the event another process attempts to create the same file (or similar) while this program has a file descriptor open on it; it won't be blocked.
- Reading from a file as it is being truncated in place is a challenge. `tail -n X -f FILE` doesn't catch this all the time; it is based on a timer. One can observe this by setting a timer delay `tail -n X -s TIME -f FILE` and truncating a file in place by replacing to the same contents. Ex. a file contains the string "abc"; replace while tail is watching with "def"; it won't catch the truncation every time. For this situation; I am relying on the following (not 100% reliable)
- If reading in chunks; the entire request is served by one read pass. File system calls on an OS on the same disk is a serial operation.
- If reading in separate chunks; (very large request); it is unlikely a truncation will happen followed by a write to the same exact spot for the next read. A seek/read error will indicate some truncation has taken place.
-- From the current position on disk; seek back 100 for a read operation; a read of 100 should work, otherwise truncation has happened.

A more foolproof way of handling truncation would possibly be a file-system watcher on the entire directory in service. If a file changes, any requests handling that file would have information ready (did the file size get smaller?). Another way could be to copy any files on a request; this would expensive.

# approach
I initially decided upon a naive way of handling reverse file reading; seek back one at a time, read back one at a time. Find new lines and read forwards from there for one line. I then implemented the various other bits of the system.

Disk I/O is best achieved by reading in large chunks; so I decided on blindly going back and reading a chunk every time was more efficient. The edge cases to solve would be partial lines at boundaries between 2 adjacent chunks. Reading backwards blindly in chunks, quickly figuring out the number of lines and processing the request in go-routines; to enable better efficiency when reading from disk to memory and allow the process of the data to be a cpu-bound problem.

I decided to not attempt to resolve the issues of:
- limited file descriptors (this can be increased via changing system configuration)
- caching requests on a file (to an extent, the linux disk cache can handle this; furthermore, caching requests can be non-optimal if the requests on the server are random).

## (chunk reading)
I spent a lot of time implementing it "righter"; I think there are still edge cases to be had.
Each chunk has a [prefix, main, suffix]; a prefix always has a valid line, and so does main, a suffix may not have a valid line.

Assuming we are reading a chunk on "123\n456\n789\nabc"; with the chunk reading the entire string at once:
- [ prefix: ["123\n"], main: ["456\n789\n"], suffix: ["abc] ]; mainCount is 2 lines

Processing this chunk, we would only process main. If we happen to now be at the start of the file, we may process prefix, since that is a guaranteed new line.

To keep it short, here are some examples:
- "123\n456\n78", chunk is 2
  - [nil, nil, 78] --> no previous block to stitch; no main to process.
  - [6\n, nil, nil] --> stitch previous [6\n, nil, nil] (discard the suffix of the previous, discard a partial line); no main to process
  - [nil, nil, 45] --> stitch previous [456\n, nil, nil]; no main to process
  - [3\n, nil, nil] --> stich previous [3\n, 456\n, nil]; main to process
  - [12, nil, nil] --> stitch previous [123\n, nil, nil]; no main to process
  - end reached, special rules to process prefix "123\n" perhaps.
  - 123 and 456 are processed

- "123\n456\n78", chunk is 5
  - [56\n, nil, 78] --> no previous block to stitch; no main to process
  - [123\n, nil, 4] --> stitch previous [123\n, 456\n, nil], main to p rocess
  - end reached, special rules to process prefix "123\n" perhaps.
  - 123 and 456 are processed

- "123\n4\n56", chunk is 5
  - [\n, 4\n, 56] --> no previous block to stitch; main to process
  - [nil, nil, 123] --> stitch previous [nil, 123\n, nil]; main to process
  - end reached, 123 and 4 are processed

Holes in the design:
- REST api chaining lines and filters is hardcoded to lines then filters; there should be a better api.
- REST api may not necessarily be rest.
- REST api response may be unnessessarily large; may require pagination. Not implemented. A problem with this is that future requests from the client may not be valid anymore due to the file continuously increasing, or being truncated/moved. We could store off a copy of this file someplace with a timeout limit for cleanup, and associate this with a token we send back to the client.
- REST api/code can open a binary file and hang.
- Golang http server code serves reach request in a go-routine; I am unsure as of now if this go thread is actually killed off when the write timeout happens; if not, we may have zombie go-routines running on forever file i/o requests.
- Related to above, possibly dealing with zombie go routines.
- For each level of the code (core -> file -> http); errors should ideally be wrapped with errors at the current abstraction level. Furthermore, the error codes should be wrapped in a way so that any error detected at the http layer doesn't just default to 404 all the time.
- How the garbage collector/memory holds up over high request periods. A pool can be written later on if required to handle the problem of reallocating memory on the heap over and over again. Would probably need some routines to shrink back down memory after a period of time if required.
- I am fairly sure the code as-is is not 100% go pedantic.
- Instead of writng to a slice or buffer and returning; it would probably be better to write to a io.Writer interface or similar; which http.ResponseWriter would also meet. There are may be optimizations (needs research) to stream the http response out vs writing it out in one huge chunk and then writing it again.

# testing
Each level of the application has unit-tests associated with it. At the file_reader level; some of the tests require the existence of 'syslog_large' in the files directory; this is not included due to file-size limits on github. This file needs to have at least 100,000 lines.

