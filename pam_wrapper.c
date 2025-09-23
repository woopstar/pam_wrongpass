// pam_wrapper.c
#include <security/pam_appl.h>
#include <security/pam_modules.h>

// Go-eksporterede symboler (cgo genererer dem)
extern int pam_sm_authenticate_go(pam_handle_t *pamh, int flags, int argc, char **argv);
extern int pam_sm_acct_mgmt_go(pam_handle_t *pamh, int flags, int argc, char **argv);

// Hjælper bruges af Go-koden
const char* get_pam_user(pam_handle_t *pamh) {
    const char *user = NULL;
    if (pam_get_user(pamh, &user, NULL) != PAM_SUCCESS || user == NULL) {
        return NULL;
    }
    return user;
}

// PAM entry points med const char **argv
int pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    return pam_sm_authenticate_go(pamh, flags, argc, (char**)argv);
}

int pam_sm_acct_mgmt(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    return pam_sm_acct_mgmt_go(pamh, flags, argc, (char**)argv);
}
