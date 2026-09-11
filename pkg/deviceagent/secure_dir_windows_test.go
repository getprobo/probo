// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:build windows

package deviceagent

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestIsTrustedWindowsSID(t *testing.T) {
	t.Parallel()

	systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	require.NoError(t, err)
	adminSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	require.NoError(t, err)
	usersSID, err := windows.CreateWellKnownSid(windows.WinBuiltinUsersSid)
	require.NoError(t, err)
	worldSID, err := windows.CreateWellKnownSid(windows.WinWorldSid)
	require.NoError(t, err)

	assert.True(t, isTrustedWindowsSID(systemSID))
	assert.True(t, isTrustedWindowsSID(adminSID))
	assert.False(t, isTrustedWindowsSID(usersSID))
	assert.False(t, isTrustedWindowsSID(worldSID))
}

func TestIsPathUnderWindowsRoot(t *testing.T) {
	t.Parallel()

	root := `D:\Data\Probo`

	assert.True(t, isPathUnderWindowsRoot(`D:\Data\Probo`, root))
	assert.True(t, isPathUnderWindowsRoot(`D:\Data\Probo\run\enrolled`, root))
	assert.True(t, isPathUnderWindowsRoot(`d:\data\probo\agent`, root))
	assert.False(t, isPathUnderWindowsRoot(`D:\Data`, root))
	assert.False(t, isPathUnderWindowsRoot(`D:\Data\Probo2`, root))
	assert.False(t, isPathUnderWindowsRoot(`C:\Temp`, root))
}

func TestEnsureWindowsProtectedDir_CreatesProtectedDACL(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "run")
	require.NoError(t, ensureWindowsProtectedDir(dir, secureDirRun))

	sids, err := daclSIDSet(dir)
	require.NoError(t, err)

	systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	require.NoError(t, err)
	adminSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	require.NoError(t, err)
	usersSID, err := windows.CreateWellKnownSid(windows.WinBuiltinUsersSid)
	require.NoError(t, err)

	assert.True(t, sids[systemSID.String()])
	assert.True(t, sids[adminSID.String()])
	assert.True(t, sids[usersSID.String()])
	assert.True(t, daclIsProtected(t, dir))
}

func TestEnsureWindowsProtectedDir_AgentOmitsUsers(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "agent")
	require.NoError(t, ensureWindowsProtectedDir(dir, secureDirAgent))

	sids, err := daclSIDSet(dir)
	require.NoError(t, err)

	usersSID, err := windows.CreateWellKnownSid(windows.WinBuiltinUsersSid)
	require.NoError(t, err)

	assert.False(t, sids[usersSID.String()])
	assert.True(t, daclIsProtected(t, dir))
}

func TestEnsureWindowsProtectedDir_RejectsUntrustedOwner(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "probo")
	require.NoError(t, os.Mkdir(dir, 0o755))

	trusted, err := isTrustedWindowsOwner(dir)
	require.NoError(t, err)

	err = ensureWindowsProtectedDir(dir, secureDirAgent)
	if trusted {
		require.NoError(t, err)
		return
	}

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not owned by SYSTEM or Administrators")
}

func TestEnsureWindowsProtectedDir_RejectsReparsePoint(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "target")
	require.NoError(t, os.Mkdir(target, 0o755))

	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}

	err := ensureWindowsProtectedDir(link, secureDirRun)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reparse point")
}

func TestIsTrustedWindowsOwner_CurrentUserFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "marker")
	require.NoError(t, os.WriteFile(path, []byte("ok\n"), 0o600))

	trusted, err := isTrustedWindowsOwner(path)
	require.NoError(t, err)

	sd, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION,
	)
	require.NoError(t, err)
	require.NotNil(t, sd)

	owner, _, err := sd.Owner()
	require.NoError(t, err)
	require.NotNil(t, owner)
	assert.Equal(t, isTrustedWindowsSID(owner), trusted)
}

func daclIsProtected(t *testing.T, path string) bool {
	t.Helper()

	sd, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
	)
	require.NoError(t, err)
	require.NotNil(t, sd)

	control, _, err := sd.Control()
	require.NoError(t, err)

	return control&windows.SE_DACL_PROTECTED != 0
}

func daclSIDSet(path string) (map[string]bool, error) {
	sd, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return nil, err
	}

	dacl, _, err := sd.DACL()
	if err != nil {
		return nil, err
	}

	found := map[string]bool{}
	for i := uint16(0); i < dacl.AceCount; i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, uint32(i), &ace); err != nil {
			return nil, err
		}

		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		found[sid.String()] = true
	}

	return found, nil
}
