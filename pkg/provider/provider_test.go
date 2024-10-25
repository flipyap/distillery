package provider_test

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/distillery/pkg/asset"
	"github.com/ekristen/distillery/pkg/osconfig"
	"github.com/ekristen/distillery/pkg/provider"
)

func init() {
	logrus.SetLevel(logrus.TraceLevel)
}

type testSourceDiscoverTest struct {
	name      string
	filenames []string
	matrix    []testSourceDiscoverMatrix
}

type testSourceDiscoverMatrix struct {
	os       string
	arch     string
	expected testSourceDiscoverExpected
}

type testSourceDiscoverExpected struct {
	error     string
	binary    string
	signature string
	checksum  string
}

func TestSourceDiscover(t *testing.T) {
	cases := []testSourceDiscoverTest{
		{
			name: "pulumi",
			filenames: []string{
				"B3SUMS",
				"B3SUMS.sig",
				"pulumi-3.133.0-checksums.txt",
				"pulumi-3.133.0-checksums.txt.sig",
				"pulumi-v3.133.0-darwin-arm64.tar.gz",
				"pulumi-v3.133.0-darwin-arm64.tar.gz.sig",
				"pulumi-v3.133.0-darwin-x64.tar.gz",
				"pulumi-v3.133.0-darwin-x64.tar.gz.sig",
				"pulumi-v3.133.0-linux-arm64.tar.gz",
				"pulumi-v3.133.0-linux-arm64.tar.gz.sig",
				"pulumi-v3.133.0-linux-x64.tar.gz",
				"pulumi-v3.133.0-linux-x64.tar.gz.sig",
				"pulumi-v3.133.0-windows-arm64.zip",
				"pulumi-v3.133.0-windows-arm64.zip.sig",
				"pulumi-v3.133.0-windows-x64.zip",
				"pulumi-v3.133.0-windows-x64.zip.sig",
				"sdk-nodejs-pulumi-pulumi-3.133.0.tgz",
				"sdk-nodejs-pulumi-pulumi-3.133.0.tgz.sig",
				"sdk-python-pulumi-3.133.0-py3-none-any.whl",
				"sdk-python-pulumi-3.133.0-py3-none-any.whl.sig",
				"SHA512SUMS",
				"SHA512SUMS.sig",
			},
			matrix: []testSourceDiscoverMatrix{ //nolint:dupl
				{
					os:   "darwin",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary:    "pulumi-v3.133.0-darwin-x64.tar.gz",
						signature: "pulumi-v3.133.0-darwin-x64.tar.gz.sig",
						checksum:  "pulumi-3.133.0-checksums.txt",
					},
				},
				{
					os:   "darwin",
					arch: "arm64",
					expected: testSourceDiscoverExpected{
						binary:    "pulumi-v3.133.0-darwin-arm64.tar.gz",
						signature: "pulumi-v3.133.0-darwin-arm64.tar.gz.sig",
						checksum:  "pulumi-3.133.0-checksums.txt",
					},
				},
				{
					os:   "linux",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary:    "pulumi-v3.133.0-linux-x64.tar.gz",
						signature: "pulumi-v3.133.0-linux-x64.tar.gz.sig",
						checksum:  "pulumi-3.133.0-checksums.txt",
					},
				},
				{
					os:   "linux",
					arch: "arm64",
					expected: testSourceDiscoverExpected{
						binary:    "pulumi-v3.133.0-linux-arm64.tar.gz",
						signature: "pulumi-v3.133.0-linux-arm64.tar.gz.sig",
						checksum:  "pulumi-3.133.0-checksums.txt",
					},
				},
				{
					os:   "windows",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary:    "pulumi-v3.133.0-windows-x64.zip",
						signature: "pulumi-v3.133.0-windows-x64.zip.sig",
						checksum:  "pulumi-3.133.0-checksums.txt",
					},
				},
			},
		},
		{
			name: "cosign",
			filenames: []string{
				"cosign-2.4.0-1.aarch64.rpm",
				"cosign-2.4.0-1.aarch64.rpm-keyless.pem",
				"cosign-2.4.0-1.aarch64.rpm-keyless.sig",
				"cosign-2.4.0-1.armv7hl.rpm",
				"cosign-2.4.0-1.armv7hl.rpm-keyless.pem",
				"cosign-2.4.0-1.armv7hl.rpm-keyless.sig",
				"cosign-2.4.0-1.ppc64le.rpm",
				"cosign-2.4.0-1.ppc64le.rpm-keyless.pem",
				"cosign-2.4.0-1.ppc64le.rpm-keyless.sig",
				"cosign-2.4.0-1.riscv64.rpm",
				"cosign-2.4.0-1.riscv64.rpm-keyless.pem",
				"cosign-2.4.0-1.riscv64.rpm-keyless.sig",
				"cosign-2.4.0-1.s390x.rpm",
				"cosign-2.4.0-1.s390x.rpm-keyless.pem",
				"cosign-2.4.0-1.s390x.rpm-keyless.sig",
				"cosign-2.4.0-1.x86_64.rpm",
				"cosign-2.4.0-1.x86_64.rpm-keyless.pem",
				"cosign-2.4.0-1.x86_64.rpm-keyless.sig",
				"cosign-darwin-amd64",
				"cosign-darwin-amd64-keyless.pem",
				"cosign-darwin-amd64-keyless.sig",
				"cosign-darwin-amd64.sig",
				"cosign-darwin-amd64_2.4.0_darwin_amd64.sbom.json",
				"cosign-darwin-arm64",
				"cosign-darwin-arm64-keyless.pem",
				"cosign-darwin-arm64-keyless.sig",
				"cosign-darwin-arm64.sig",
				"cosign-darwin-arm64_2.4.0_darwin_arm64.sbom.json",
				"cosign-linux-amd64",
				"cosign-linux-amd64-keyless.pem",
				"cosign-linux-amd64-keyless.sig",
				"cosign-linux-amd64.sig",
				"cosign-linux-amd64_2.4.0_linux_amd64.sbom.json",
				"cosign-linux-arm",
				"cosign-linux-arm-keyless.pem",
				"cosign-linux-arm-keyless.sig",
				"cosign-linux-arm.sig",
				"cosign-linux-arm64",
				"cosign-linux-arm64-keyless.pem",
				"cosign-linux-arm64-keyless.sig",
				"cosign-linux-arm64.sig",
				"cosign-linux-arm64_2.4.0_linux_arm64.sbom.json",
				"cosign-linux-arm_2.4.0_linux_arm.sbom.json",
				"cosign-linux-pivkey-pkcs11key-amd64",
				"cosign-linux-pivkey-pkcs11key-amd64-keyless.pem",
				"cosign-linux-pivkey-pkcs11key-amd64-keyless.sig",
				"cosign-linux-pivkey-pkcs11key-amd64.sig",
				"cosign-linux-pivkey-pkcs11key-amd64_2.4.0_linux_amd64.sbom.json",
				"cosign-linux-pivkey-pkcs11key-arm64",
				"cosign-linux-pivkey-pkcs11key-arm64-keyless.pem",
				"cosign-linux-pivkey-pkcs11key-arm64-keyless.sig",
				"cosign-linux-pivkey-pkcs11key-arm64.sig",
				"cosign-linux-pivkey-pkcs11key-arm64_2.4.0_linux_arm64.sbom.json",
				"cosign-linux-ppc64le",
				"cosign-linux-ppc64le-keyless.pem",
				"cosign-linux-ppc64le-keyless.sig",
				"cosign-linux-ppc64le.sig",
				"cosign-linux-ppc64le_2.4.0_linux_ppc64le.sbom.json",
				"cosign-linux-riscv64",
				"cosign-linux-riscv64-keyless.pem",
				"cosign-linux-riscv64-keyless.sig",
				"cosign-linux-riscv64.sig",
				"cosign-linux-riscv64_2.4.0_linux_riscv64.sbom.json",
				"cosign-linux-s390x",
				"cosign-linux-s390x-keyless.pem",
				"cosign-linux-s390x-keyless.sig",
				"cosign-linux-s390x.sig",
				"cosign-linux-s390x_2.4.0_linux_s390x.sbom.json",
				"cosign-windows-amd64.exe",
				"cosign-windows-amd64.exe-keyless.pem",
				"cosign-windows-amd64.exe-keyless.sig",
				"cosign-windows-amd64.exe.sig",
				"cosign-windows-amd64.exe_2.4.0_windows_amd64.sbom.json",
				"cosign_2.4.0_aarch64.apk",
				"cosign_2.4.0_aarch64.apk-keyless.pem",
				"cosign_2.4.0_aarch64.apk-keyless.sig",
				"cosign_2.4.0_amd64.deb",
				"cosign_2.4.0_amd64.deb-keyless.pem",
				"cosign_2.4.0_amd64.deb-keyless.sig",
				"cosign_2.4.0_arm64.deb",
				"cosign_2.4.0_arm64.deb-keyless.pem",
				"cosign_2.4.0_arm64.deb-keyless.sig",
				"cosign_2.4.0_armhf.deb",
				"cosign_2.4.0_armhf.deb-keyless.pem",
				"cosign_2.4.0_armhf.deb-keyless.sig",
				"cosign_2.4.0_armv7.apk",
				"cosign_2.4.0_armv7.apk-keyless.pem",
				"cosign_2.4.0_armv7.apk-keyless.sig",
				"cosign_2.4.0_ppc64el.deb",
				"cosign_2.4.0_ppc64el.deb-keyless.pem",
				"cosign_2.4.0_ppc64el.deb-keyless.sig",
				"cosign_2.4.0_ppc64le.apk",
				"cosign_2.4.0_ppc64le.apk-keyless.pem",
				"cosign_2.4.0_ppc64le.apk-keyless.sig",
				"cosign_2.4.0_riscv64.apk",
				"cosign_2.4.0_riscv64.apk-keyless.pem",
				"cosign_2.4.0_riscv64.apk-keyless.sig",
				"cosign_2.4.0_riscv64.deb",
				"cosign_2.4.0_riscv64.deb-keyless.pem",
				"cosign_2.4.0_riscv64.deb-keyless.sig",
				"cosign_2.4.0_s390x.apk",
				"cosign_2.4.0_s390x.apk-keyless.pem",
				"cosign_2.4.0_s390x.apk-keyless.sig",
				"cosign_2.4.0_s390x.deb",
				"cosign_2.4.0_s390x.deb-keyless.pem",
				"cosign_2.4.0_s390x.deb-keyless.sig",
				"cosign_2.4.0_x86_64.apk",
				"cosign_2.4.0_x86_64.apk-keyless.pem",
				"cosign_2.4.0_x86_64.apk-keyless.sig",
				"cosign_checksums.txt",
				"cosign_checksums.txt-keyless.pem",
				"cosign_checksums.txt-keyless.sig",
				"release-cosign.pub",
			},
			matrix: []testSourceDiscoverMatrix{ //nolint:dupl
				{
					os:   "darwin",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary:    "cosign-darwin-amd64",
						signature: "cosign-darwin-amd64.sig",
						checksum:  "cosign_checksums.txt",
					},
				},
				{
					os:   "darwin",
					arch: "arm64",
					expected: testSourceDiscoverExpected{
						binary:    "cosign-darwin-arm64",
						signature: "cosign-darwin-arm64.sig",
						checksum:  "cosign_checksums.txt",
					},
				},
				{
					os:   "linux",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary:    "cosign-linux-amd64",
						signature: "cosign-linux-amd64.sig",
						checksum:  "cosign_checksums.txt",
					},
				},
				{
					os:   "linux",
					arch: "arm64",
					expected: testSourceDiscoverExpected{
						binary:    "cosign-linux-arm64",
						signature: "cosign-linux-arm64.sig",
						checksum:  "cosign_checksums.txt",
					},
				},
				{
					os:   "windows",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary:    "cosign-windows-amd64.exe",
						signature: "cosign-windows-amd64.exe.sig",
						checksum:  "cosign_checksums.txt",
					},
				},
			},
		},
		{
			name: "acorn",
			filenames: []string{
				"acorn-v0.10.1-linux-amd64.tar.gz",
				"acorn-v0.10.1-linux-arm64.tar.gz",
				"acorn-v0.10.1-macOS-universal.tar.gz",
				"acorn-v0.10.1-macOS-universal.zip",
				"acorn-v0.10.1-windows-amd64.zip",
			},
			matrix: []testSourceDiscoverMatrix{
				{
					os:   "darwin",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary: "acorn-v0.10.1-macOS-universal.tar.gz",
					},
				},
				{
					os:   "darwin",
					arch: "arm64",
					expected: testSourceDiscoverExpected{
						binary: "acorn-v0.10.1-macOS-universal.tar.gz",
					},
				},
				{
					os:   "linux",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary:    "acorn-v0.10.1-linux-amd64.tar.gz",
						signature: "",
						checksum:  "",
					},
				},
				{
					os:   "linux",
					arch: "arm64",
					expected: testSourceDiscoverExpected{
						binary:    "acorn-v0.10.1-linux-arm64.tar.gz",
						signature: "",
						checksum:  "",
					},
				},
				{
					os:   "windows",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						binary:    "acorn-v0.10.1-windows-amd64.zip",
						signature: "",
						checksum:  "",
					},
				},
			},
		},
		{
			name: "nerdctl",
			filenames: []string{
				"nerdctl-1.7.7-freebsd-amd64.tar.gz",
				"nerdctl-1.7.7-go-mod-vendor.tar.gz",
				"nerdctl-1.7.7-linux-amd64.tar.gz",
				"nerdctl-1.7.7-linux-amd-v7.tar.gz",
				"nerdctl-1.7.7-linux-arm64.tar.gz",
				"nerdctl-1.7.7-linux-ppc64le.tar.gz",
				"nerdctl-1.7.7-linux-riscv64.tar.gz",
				"nerdctl-1.7.7-linux-s390x.tar.gz",
				"nerdctl-1.7.7-windows-amd64.tar.gz",
				"nerdctl-full-1.7.7-linux-amd64.tar.gz",
				"nerdctl-full-1.7.7-linux-arm64.tar.gz",
				"SHA256SUMS",
				"SHA256SUMS.asc",
			},
			matrix: []testSourceDiscoverMatrix{
				{
					os:   "darwin",
					arch: "amd64",
					expected: testSourceDiscoverExpected{
						error:     "no matching asset found, score too low",
						binary:    "",
						signature: "",
						checksum:  "",
					},
				},
				{
					os:   "linux",
					arch: "arm64",
					expected: testSourceDiscoverExpected{
						binary:    "nerdctl-1.7.7-linux-arm64.tar.gz",
						signature: "SHA256SUMS.asc",
						checksum:  "SHA256SUMS",
					},
				},
			},
		},
	}

	t.Parallel()
	for _, tc := range cases {
		for _, m := range tc.matrix {
			t.Run(fmt.Sprintf("%s-%s-%s", tc.name, m.os, m.arch), func(t *testing.T) {
				var assets []asset.IAsset
				for _, filename := range tc.filenames {
					newA := &asset.Asset{
						Name:        filename,
						DisplayName: filename,
						OS:          m.os,
						Arch:        m.arch,
					}
					newA.Classify()
					assets = append(assets, newA)
				}

				testSource := provider.Provider{
					OSConfig: osconfig.New(m.os, m.arch),
					Options: &provider.Options{
						OS:   m.os,
						Arch: m.arch,
						Settings: map[string]interface{}{
							"no-score-check": false,
						},
					},
					Assets: assets,
				}

				err := testSource.Discover([]string{tc.name})
				if m.expected.error != "" {
					assert.EqualError(t, err, m.expected.error)
					return
				}

				assert.NoError(t, err)

				if m.expected.binary != "" {
					assert.Equal(t, m.expected.binary, testSource.Binary.GetName(), "expected binary")
				}
				if m.expected.signature != "" {
					assert.Equal(t, m.expected.signature, testSource.Signature.GetName(), "expected signature")
				}
				if m.expected.checksum != "" {
					assert.Equal(t, m.expected.checksum, testSource.Checksum.GetName(), "expected checksum")
				}
			})
		}
	}
}
