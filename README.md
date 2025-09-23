# PAM Wrongpass

A simple PAM module written in Go that counts failed login attempts and takes action once a defined threshold is reached.
By default, the module **shuts down the system** (`systemctl poweroff`) when too many incorrect login attempts are detected.

---

## Features

- Counts failed login attempts per user.
- Securely stores the counter in `/var/lib/pam_wrongpass`.
- Resets the counter on successful login.
- Shuts down the system when the number of failed attempts reaches `max_tries`.
- `max_tries` can be configured directly in the PAM configuration.

---

## Installation

### 1. Download the latest release
Go to [Releases](../../releases) and download the `.so` file for your architecture:

- `pam_wrongpass_linux_amd64.so` – 64-bit x86 (most desktops and servers)
- `pam_wrongpass_linux_arm64.so` – 64-bit ARM (e.g., Raspberry Pi)

Copy the file to the PAM modules directory:

```bash
sudo install -m 0644 pam_wrongpass_linux_amd64.so /usr/lib/security/pam_wrongpass.so
````

> **Note:** Some distributions use `/lib/security/` instead of `/usr/lib/security/`.

---

### 2. Create the state directory

The module stores counters securely in `/var/lib/pam_wrongpass`.
Make sure the directory exists and has the correct permissions:

```bash
sudo mkdir -p /var/lib/pam_wrongpass
sudo chmod 700 /var/lib/pam_wrongpass
sudo chown root:root /var/lib/pam_wrongpass
```

---

### 3. Configure PAM

Edit your system's PAM configuration files for login.
On most systems, these are `common-auth` and `common-account` under `/etc/pam.d/`.

Add **at the top** of `/etc/pam.d/common-auth`:

```
auth    optional    pam_wrongpass.so max_tries=5
```

Add **at the top** of `/etc/pam.d/common-account`:

```
account required    pam_wrongpass.so
```

* `max_tries=5` means the system will shut down after **5 failed login attempts**.
* If omitted, the default value is `10`.

---

## Testing the module

You can test using [pamtester](https://github.com/firelizzard/pamtester):

```bash
sudo pamtester login <your_username> authenticate
```

Enter incorrect passwords repeatedly.
When the threshold is reached, the system will start shutting down via `systemctl poweroff`.

---

## Build from source

### 1. Install dependencies

On Debian/Ubuntu:

```bash
sudo apt-get update
sudo apt-get install -y build-essential libpam0g-dev golang
```

### 2. Build the `.so` file

```bash
go build -buildmode=c-shared -o pam_wrongpass.so pam_wrongpass.go
```

### 3. Install

Copy the built file into the PAM directory:

```bash
sudo install -m 0644 pam_wrongpass.so /usr/lib/security/pam_wrongpass.so
```

---

## GitHub Actions CI/CD

This repository includes GitHub Actions workflows to build the module for multiple architectures:

* **amd64** – built natively
* **arm64** – built using QEMU and [`uraimo/run-on-arch-action`](https://github.com/uraimo/run-on-arch-action)

When you push a **tag** (e.g., `v1.0.0`):

1. GitHub Actions automatically builds `.so` files for all supported architectures.
2. The build artifacts are uploaded to the GitHub Release.

### Example

```bash
git tag v1.0.0
git push origin v1.0.0
```

The release files will then appear under [Releases](../../releases).

---

## Configuration

| Parameter   | Description                                     | Default |
| ----------- | ----------------------------------------------- | ------- |
| `max_tries` | Number of failed login attempts before shutdown | `10`    |

Example configuration with shutdown after **3 failed attempts**:

```
auth    optional    pam_wrongpass.so max_tries=3
```

---

## Security Considerations

* Counters are stored securely in `/var/lib/pam_wrongpass` with `root:root` ownership and `0700` permissions to prevent tampering.
* Users cannot reset or modify the counter themselves.
* If you want to **lock the system** instead of shutting it down, replace the `systemctl poweroff` command in the source code with another action, such as disabling the user account.

---

## Troubleshooting

* Check PAM logs:

  ```bash
  sudo journalctl -xe | grep pam_wrongpass
  ```
* Ensure the `.so` file is in the correct PAM directory:

  * `/usr/lib/security/` or `/lib/security/` depending on your distribution.
* Verify that `libpam0g-dev` is installed if building from source.

---

## Development

### Run locally

If you need to iterate during development, you can run the build workflow manually using:

```bash
make build
```

Or trigger the GitHub Action using:

```bash
gh workflow run "Build PAM Wrongpass Module"
```

---

## Disclaimer

Use at your own risk.
Misconfiguring a PAM module can lock you out of your system.
**Always test in a safe environment** (such as a virtual machine) before deploying on production systems.

---

## License

This project is released under the [MIT License](LICENSE).

```
