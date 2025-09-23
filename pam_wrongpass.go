// pam_wrongpass.go
package main

/*
#cgo CFLAGS: -fPIC
#cgo LDFLAGS: -lpam
#include <security/pam_appl.h>
#include <security/pam_modules.h>

// C-funktioner implementeret i pam_wrapper.c (kun declarations her)
int pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv);
int pam_sm_acct_mgmt(pam_handle_t *pamh, int flags, int argc, const char **argv);

// Helper fra pam_wrapper.c (kun declaration)
const char* get_pam_user(pam_handle_t *pamh);
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
	_ = os.MkdirAll(stateDir, 0700)
	_ = exec.Command("/usr/bin/systemctl", "poweroff").Start()
}

func parseMaxTries(argc C.int, argv **C.char) int {
	maxTries := defaultMaxTries
	if argc <= 0 || argv == nil {
		return maxTries
	}
	args := (*[1 << 16]*C.char)(unsafe.Pointer(argv))[:argc:argc]
	for _, a := range args {
		if a == nil {
			continue
		}
		s := C.GoString(a)
		if strings.HasPrefix(s, "max_tries=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(s, "max_tries=")); err == nil && v > 0 {
				maxTries = v
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
	cnt := readInt(path) + 1
	writeInt(path, cnt)

	if cnt >= maxTries {
		writeInt(path, 0)
		poweroffAsync()
	}
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
