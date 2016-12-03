// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var pkgOrder = []string{
	// L0 Packages
	"unsafe",
	"runtime/internal/sys",
	"runtime/internal/atomic",
	"runtime",
	"sync/atomic",
	"internal/race",
	"sync",
	"errors",
	"io",

	// L1 Packages
	"unicode/utf8",
	"unicode/utf16",
	"sort",
	"math",
	"math/cmplx",
	"math/rand",
	"strconv",

	// L2 Packages
	"unicode",
	"strings",
	"bytes",
	"path",
	"bufio",

	// L3 Packages
	"crypto/subtle",
	"reflect",
	"encoding/base32",
	"encoding/base64",
	"encoding/binary",
	"hash",
	"hash/adler32",
	"hash/crc32",
	"hash/crc64",
	"hash/fnv",
	"crypto",
	"crypto/cipher",
	"image/color",
	"image",
	"image/color/palette",

	// Operating system access
	"internal/syscall/windows/sysdll",
	"syscall",
	"internal/syscall/unix",
	//"internal/syscall/windows",
	//"internal/syscall/windows/registry",
	"time",
	"os",
	"path/filepath",
	"io/ioutil",
	"os/signal",
	"fmt",
	"log",
	"context",
	"os/exec",

	// Low level testing dependencies
	"regexp/syntax",
	"regexp",
	"text/tabwriter",
	"runtime/debug",
	"runtime/pprof",
	"runtime/trace",
	"flag",
	"testing",
	"testing/iotest",
	"testing/quick",
	"internal/testenv",

	// Go parser
	"go/token",
	"go/scanner",
	"go/ast",
	"go/parser",
	"go/printer",
	"text/template/parse",
	"net/url",
	"text/template",
	"go/doc",
	"go/format",

	// Go type checking
	"math/big",
	"go/constant",
	"go/build",
	"container/heap",
	"go/types",
	"text/scanner",
	"compress/flate",
	"compress/zlib",
	"debug/dwarf",
	"debug/elf",
	"go/internal/gcimporter",
	"go/internal/gccgoimporter",
	"go/importer",

	// One of a kind
	"archive/tar",
	"archive/zip",
	"compress/bzip2",
	"compress/gzip",
	"compress/lzw",
	"container/list",
	"database/sql/driver",
	"database/sql",
	"debug/gosym",
	"debug/macho",
	"debug/pe",
	"debug/plan9obj",
	"encoding",
	"encoding/ascii85",
	"encoding/asn1",
	"encoding/csv",
	"encoding/gob",
	"encoding/hex",
	"encoding/json",
	"encoding/pem",
	"encoding/xml",
	"html",
	"image/internal/imageutil",
	"image/draw",
	"image/gif",
	"image/jpeg",
	"image/png",
	"index/suffixarray",
	"internal/singleflight",
	"internal/trace",
	"mime",
	"mime/quotedprintable",
	"net/internal/socktest",
	"html/template",

	// CGO related
	"runtime/cgo",
	"runtime/race",
	//"runtime/msan",
	"os/user",

	// Basic networking
	"internal/nettrace",
	"net",

	// Uses of networking
	"net/textproto",
	"net/mail",
	"log/syslog", // Panic in linker...

	// Core crypto
	"crypto/aes",
	"crypto/des",
	"crypto/hmac",
	"crypto/md5",
	"crypto/rc4",
	"crypto/sha1",
	"crypto/sha256",
	"crypto/sha512",

	// Crypto random
	"crypto/rand",

	// Mathematical crypto
	"crypto/rsa",
	"crypto/elliptic",
	"crypto/ecdsa", // Panic in linker...
	"crypto/dsa",

	// SSL/TLS
	"crypto/x509/pkix",
	//"crypto/x509",
	//"crypto/tls",

	// net + crypto
	"mime/multipart",
	//"net/smtp",

	// HTTP
	"net/http/httptrace",
	"net/http/internal",
	//"http", // Depends on crypto/tls

	// HTTP-using packages (TODO - depends on /net/http)
	//"expvar",
	//"net/http/cgi",
	//"net/http/cookiejar",
	//"net/http/fcgi",
	//"net/http/httptest",
	//"net/http/httputil",
	//"net/http/pprof",
	//"net/rpc",
	//"net/rpc/jsonrpc",
}

var ldflags = flag.String("ldflags", "", "The flags to pass on to the linker")

func main() {
	// Parse command line flags
	flag.Parse()

	// First compile the standard library as one giant package
	log.Println("Building libstd.so")
	err := compilestd()
	if err != nil {
		log.Fatalln("Failed to compile the golang standard library libstd.so:", err)
	}
	log.Println("Successfully built libstd.so")
	time.Sleep(2 * time.Second)

	// Walk through the packages in dependency order and attemp to build
	for _, pkg := range pkgOrder {
		log.Println("Building package:", pkg)
		err := trycompile(pkg)
		if err != nil {
			log.Println("Error building package:", err)
		} else {
			log.Println("Successfully Built:", pkg)
		}

		// Give the filesystem a little time
		time.Sleep(100 * time.Millisecond)
	}
}

func compilestd() error {
	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	return runinstall(ctxt, []string{"std"}, filepath.Join(os.Getenv("GOROOT"), "src"))
}

func trycompile(pkg string) error {
	ctxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return runinstall(ctxt, nil, filepath.Join(os.Getenv("GOROOT"), "src", pkg))
}

func runinstall(ctxt context.Context, extraargs []string, dir string) error {
	// Setup the command parameters
	args := []string{"install", "-ldflags", *ldflags, "-buildmode", "shared", "-linkshared", "-v"}
	args = append(args, extraargs...)

	// Create the command
	command := exec.CommandContext(ctxt, "go", args...)
	command.Dir = dir
	command.Env = nil
	out, err := command.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			log.Println(string(out))
		}
		return err
	}

	return nil
}
