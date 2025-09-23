// pam_wrongpass.go
package main

/*
#cgo CFLAGS: -fPIC
#cgo LDFLAGS: -lpam
#include <security/pam_appl.h>
#include <security/pam_modules.h>
#include <stdlib.h>

// Forward declarations to Go (non-standard names)
int pam_sm_authenticate_go(pam_handle_t *pamh, int flags, int argc, char **argv);
int pam_sm_acct_mgmt_go(pam_handle_t *pamh, int flags, int argc, char **argv);

// Wrappers with official PAM signatures using const char **argv
int pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    // cast away const only for reading argv values
    return pam_sm_authenticate_go(pamh, flags, argc, (char**)argv);
}
int pam_sm_acct_mgmt(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    return pam_sm_acct_mgmt_go(pamh, flags, argc, (char**)argv);
}

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

func parseMaxTries(argc C.int, argv **C.char) int {
	maxTries := defaultMaxTries
	if argc <= 0 || argv == nil {
		return maxTries
	}
	// read argv
	argSlice := (*[1 << 16]*C.char)(unsafe.Pointer(argv))[:argc:argc]
	for _, argC := range argSlice {
		if argC == nil {
			continue
		}
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

//export pam_sm_authenticate_go
func pam_sm_authenticate_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
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
	// never decide auth result
	return C.PAM_SUCCESS
}

//export pam_sm_acct_mgmt_go
func pam_sm_acct_mgmt_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	userC := C.get_pam_user(pamh)
	if userC == nil {
		return C.PAM_SUCCESS
	}
	user := C.GoString(userC)
	path := countFilePathForUser(user)
	writeInt(path, 0)
	return C.PAM_SUCCESS
}

// required for -buildmode=c-shared
func main() {}
