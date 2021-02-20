package core_utils

import (
    "io"
)

func SeekEnd(buffer io.ReadSeeker) (io.ReadSeeker, error) {
    _, err := buffer.Seek(0, io.SeekEnd)
    return buffer, err
}

