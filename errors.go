package np

type Error struct {
	err string
}

func (e Error) Error() string { return e.err }

var (

	// https://github.com/0intro/plan9/blob/7524062cfa4689019a4ed6fc22500ec209522ef0/sys/src/lib9p/srv.c#L10
	ErrBadAttach    = Error{err: "unknown specifier in attach"}
	ErrBadOffset    = Error{err: "bad offset"}
	ErrBadCount     = Error{err: "bad count"}
	ErrBotch        = Error{err: "9P protocol botch"} // EPROTO
	ErrCreateNonDir = Error{err: "create in non-directory"}
	ErrDupFid       = Error{err: "duplicate fid"}
	ErrDupTag       = Error{err: "duplicate tag"}
	ErrNoCreate     = Error{err: "create prohibited"}
	ErrNoRemove     = Error{err: "remove prohibited"}
	ErrNoStat       = Error{err: "stat prohibited"}
	ErrNotFound     = Error{err: "file not found"} // ENOENT
	ErrNoWrite      = Error{err: "write prohibited"}
	ErrNoWstat      = Error{err: "wstat prohibited"}  // EPERM
	ErrPerm         = Error{err: "permission denied"} // EACCES
	ErrUnknownFid   = Error{err: "unknown fid"}
	ErrBadDir       = Error{err: "bad directory in wstat"}
	ErrWalkNoDir    = Error{err: "walk in non-directory"}

	// https://github.com/torvalds/linux/blob/6e195b0f7c8e50927fa31946369c22a0534ec7e2/net/9p/error.c#L40

	ErrBadFD    = Error{err: "File descriptor in bad state"} // EBADFD
	ErrBadFid   = Error{err: "bad use of fid"}               // EBADF
	ErrFidInUse = Error{err: "fid already in use"}           // EBADF

	ErrAuth = Error{err: "authentication failed"} // ECONNREFUSED

	ErrCrossDevice = Error{err: "Invalid cross-device link"} // EXDEV
	ErrDeadlock    = Error{err: "Resource deadlock avoided"} // EDEADLK
	ErrDirNotEmpty = Error{err: "directory is not empty"}    // ENOTEMPTY

	ErrExists = Error{err: "file exists"}  // EEXIST
	ErrInUse  = Error{err: "file in use"}  // ETXTBSY
	ErrTooBig = Error{err: "file too big"} // EFBIG

	ErrIllegalMode   = Error{err: "illegal mode"}   // EINVAL
	ErrIllegalName   = Error{err: "illegal name"}   // ENAMETOOLONG
	ErrIllegalOffset = Error{err: "illegal offset"} // EINVAL
	ErrIllegalSeek   = Error{err: "Illegal seek"}   // ESPIPE

	ErrInProgress  = Error{err: "Operation now in progress"} // EINPROGRESS
	ErrInterrupted = Error{err: "Interrupted system call"}   // EINTR
	ErrInvalidArg  = Error{err: "Invalid argument"}          // EINVAL
	ErrIO          = Error{err: "i/o error"}                 // EIO

	ErrBadMessage     = Error{err: "Bad message"}                // EBADMSG
	ErrMessageTooLong = Error{err: "Message too long"}           // EMSGSIZE
	ErrNoMessage      = Error{err: "No message of desired type"} // ENOMSG

	ErrConnAbort      = Error{err: "Software caused connection abort"}        // ECONNABORTED
	ErrConnected      = Error{err: "Transport endpoint is already connected"} // EISCONN
	ErrConnRefused    = Error{err: "Connection refused"}                      // ECONNREFUSED
	ErrConnReset      = Error{err: "Connection reset by peer"}                // ECONNRESET
	ErrHostDown       = Error{err: "Host is down"}                            // EHOSTDOWN
	ErrNetDown        = Error{err: "Network is down"}                         // ENETDOWN
	ErrNetReset       = Error{err: "Network dropped connection on reset"}     // ENETRESET
	ErrNetUnreachable = Error{err: "Network is unreachable"}                  // ENETUNREACH
	ErrNoNet          = Error{err: "Machine is not on the network"}           // ENONET
	ErrNoRoute        = Error{err: "No route to host"}                        // EHOSTUNREACH
	ErrNotConnected   = Error{err: "Transport endpoint is not connected"}     // ENOTCONN

	ErrNoDevice       = Error{err: "No such device"}            // ENODEV
	ErrNoDeviceOrAddr = Error{err: "No such device or address"} // ENXIO
	ErrNoLink         = Error{err: "Link has been severed"}     // ENOLINK
	ErrNoLock         = Error{err: "No locks available"}        // ENOLCK
	ErrNoMem          = Error{err: "Cannot allocate memory"}    // ENOMEM
	ErrNoPackage      = Error{err: "Package not installed"}     // ENOPKG

	ErrBrokenPipe    = Error{err: "Broken pipe"}                 // EPIPE
	ErrBadAddr       = Error{err: "Bad address"}                 // EFAULT
	ErrBusy          = Error{err: "Device or resource busy"}     // EBUSY
	ErrComm          = Error{err: "Communication error on send"} // ECOMM
	ErrNoBufferSpace = Error{err: "No buffer space available"}   // ENOBUFS
	ErrNoData        = Error{err: "No data available"}           // ENODATA
	ErrNoSpace       = Error{err: "file system is full"}         // ENOSPC

	ErrAllreadyInProgress = Error{err: "Operation already in progress"}                 // EALREADY
	ErrShutdown           = Error{err: "Cannot send after transport endpoint shutdown"} // ESHUTDOWN
	ErrTimeout            = Error{err: "Connection timed out"}                          // ETIMEDOUT

	ErrIsDir       = Error{err: "Is a directory"}                 // EISDIR
	ErrIsNamed     = Error{err: "Is a named type file"}           // EISNAM
	ErrNotBlockDev = Error{err: "Block device required"}          // ENOTBLK
	ErrNotDir      = Error{err: "not a directory"}                // ENOTDIR
	ErrNotSock     = Error{err: "Socket operation on non-socket"} // ENOTSOCK

	ErrNotImplemented = Error{err: "Function not implemented"} // ENOSYS
	ErrOpNoSupported  = Error{err: "Operation not supported"}  // EOPNOTSUPP

	ErrOutOfRange = Error{err: "Numerical argument out of domain"} // EDOM
	ErrQuota      = Error{err: "Disk quota exceeded"}              // EDQUOT
	ErrRange      = Error{err: "Numerical result out of range"}    // ERANGE
	ErrReadOnly   = Error{err: "file is read only"}                // EROFS
	ErrReadOnlyFS = Error{err: "read only file system"}            // EROFS
	ErrRemote     = Error{err: "Object is remote"}                 // EREMOTE
	ErrRemoteIO   = Error{err: "Remote I/O error"}                 // EREMOTEIO
	ErrRemoved    = Error{err: "file has been removed"}            // EIDRM
	ErrStreamPipe = Error{err: "Streams pipe error"}               // ESTRPIPE

	ErrNoProto              = Error{err: "Protocol not available"}        // ENOPROTOOPT
	ErrProtoNoSupport       = Error{err: "Protocol not supported"}        // EPROTONOSUPPORT
	ErrProtoFamilyNoSupport = Error{err: "Protocol family not supported"} // EPFNOSUPPORT
	ErrSockNoSupported      = Error{err: "Socket type not supported"}     // ESOCKTNOSUPPORT

	ErrTooManyArgs      = Error{err: "Argument list too long"}            // E2BIG
	ErrTooManyFiles     = Error{err: "Too many open files"}               // EMFILE
	ErrTooManyLevels    = Error{err: "Too many levels of symbolic links"} // ELOOP
	ErrTooManyLinks     = Error{err: "Too many links"}                    // EMLINK
	ErrTooManyOpenFiles = Error{err: "Too many open files in system"}     // ENFILE
	ErrTooManyUsers     = Error{err: "Too many users"}                    // EUSERS

	ErrTempUnavailable = Error{err: "Resource temporarily unavailable"} // EAGAIN

	ErrUnknownGroup    = Error{err: "unknown group"}               // EINVAL
	ErrUnknownOrBadFid = Error{err: "fid unknown or out of range"} // EBADF
	ErrUnknownUser     = Error{err: "unknown user"}                // EINVAL
)
