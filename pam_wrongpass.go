// pam_wrongpass.go
package main

/*
#cgo CFLAGS: -fPIC
#cgo LDFLAGS: -lpam
#include <security/pam_appl.h>
#include <security/pam_modules.h>
#include <stdlib.h>

static const char* get_pam_user(pam_handle_t *pamh) {
    const char *user = NULL;
    if (pam_get_user(pamh, &user, NULL) != PAM_SUCCESS || user == NULL) {
        return NULL;
    }
    return user;
}
*/
import "C"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"
)

const (
	defaultMaxTries = 10
	stateDir        = "/var/lib/pam_wrongpass"
)

// Resolve path for a given user
func countFilePathForUser(user string) string {
	return filepath.Join(stateDir, fmt.Sprintf("%s.count", user))
}

func readInt(path string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(b))
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func writeInt(path string, v int) {
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	_ = os.WriteFile(path, []byte(fmt.Sprintf("%d\n", v)), 0600)
}

func poweroffAsync() {
	_ = exec.Command("/usr/bin/systemctl", "poweroff").Start()
}

// Parse maxTries from PAM arguments
func parseMaxTries(argc C.int, argv **C.char) int {
	maxTries := defaultMaxTries
	argSlice := (*[1 << 16]*C.char)(unsafe.Pointer(argv))[:argc:argc]

	for _, argC := range argSlice {
		arg := C.GoString(argC)
		if strings.HasPrefix(arg, "max_tries=") {
			valStr := strings.TrimPrefix(arg, "max_tries=")
			if val, err := strconv.Atoi(valStr); err == nil && val > 0 {
				maxTries = val
			}
		}
	}
	return maxTries
}

// Called during authentication (always called, success or fail)
//
//export pam_sm_authenticate
func pam_sm_authenticate(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	userC := C.get_pam_user(pamh)
	if userC == nil {
		return C.PAM_SUCCESS
	}
	user := C.GoString(userC)
	maxTries := parseMaxTries(argc, argv)

	path := countFilePathForUser(user)
	cnt := readInt(path)
	cnt++

	writeInt(path, cnt)

	if cnt >= maxTries {
		writeInt(path, 0)
		poweroffAsync()
	}

	// Do not block login process
	return C.PAM_SUCCESS
}

// Called on account phase after successful authentication
//
//export pam_sm_acct_mgmt
func pam_sm_acct_mgmt(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	userC := C.get_pam_user(pamh)
	if userC == nil {
		return C.PAM_SUCCESS
	}
	user := C.GoString(userC)

	path := countFilePathForUser(user)
	writeInt(path, 0)

	return C.PAM_SUCCESS
}

// Required for buildmode=c-shared
func main() {}
