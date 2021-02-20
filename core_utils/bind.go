package core_utils

import (
    "io"
)

type LogFuncMonad func(io.ReadSeeker) (io.ReadSeeker, error)

func LogFuncBind(buffer io.ReadSeeker, err error, f ...LogFuncMonad) (io.ReadSeeker, error) {
    if err != nil {
        return nil, err
    }

    if _, err := buffer.Seek(0, io.SeekEnd); err != nil {
        return nil, err
    }
    if len(f) == 1 {
        return f[0](buffer)
    }
    res, err := f[0](buffer)
    return LogFuncBind(res, err, f[1:]...)
}
