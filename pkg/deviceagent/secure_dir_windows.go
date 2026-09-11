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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func ensureProtectedWindowsTree(path string, kind secureDirKind) error {
	if !isProtectedWindowsPath(path) {
		return nil
	}

	root := DefaultProgramDataRoot()
	if err := ensureWindowsProtectedDir(root, secureDirRoot); err != nil {
		return err
	}

	if isSameWindowsPath(path, root) {
		return nil
	}

	return ensureWindowsProtectedDir(path, kind)
}

func isTrustedEnrollmentMarker(path string) (bool, error) {
	if !isProtectedWindowsPath(path) {
		return true, nil
	}

	return isTrustedWindowsOwner(path)
}

func isProtectedWindowsPath(path string) bool {
	return isPathUnderWindowsRoot(path, DefaultProgramDataRoot())
}

func isPathUnderWindowsRoot(path, root string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}

	pathFold := strings.ToLower(filepath.Clean(absPath))
	rootFold := strings.ToLower(filepath.Clean(absRoot))
	if pathFold == rootFold {
		return true
	}

	return strings.HasPrefix(pathFold, rootFold+string(os.PathSeparator))
}

func isSameWindowsPath(a, b string) bool {
	absA, err := filepath.Abs(a)
	if err != nil {
		return false
	}

	absB, err := filepath.Abs(b)
	if err != nil {
		return false
	}

	return strings.EqualFold(filepath.Clean(absA), filepath.Clean(absB))
}

func ensureWindowsProtectedDir(path string, kind secureDirKind) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return createWindowsProtectedDir(path, kind)
		}

		return fmt.Errorf("cannot stat %s: %w", path, err)
	}

	reparse, err := isWindowsReparsePoint(path)
	if err != nil {
		return fmt.Errorf("cannot inspect %s: %w", path, err)
	}

	if reparse || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s is a reparse point; remove it and retry", path)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	trusted, err := isTrustedWindowsOwner(path)
	if err != nil {
		return fmt.Errorf("cannot check owner of %s: %w", path, err)
	}

	if !trusted {
		return fmt.Errorf(
			"%s is not owned by SYSTEM or Administrators; remove it and retry",
			path,
		)
	}

	if err := applyWindowsProtectedDACL(path, kind); err != nil {
		return fmt.Errorf("cannot apply protected DACL to %s: %w", path, err)
	}

	return nil
}

func createWindowsProtectedDir(path string, kind secureDirKind) error {
	acl, err := windowsProtectedACL(kind)
	if err != nil {
		return fmt.Errorf("cannot build DACL for %s: %w", path, err)
	}

	sd, err := windows.NewSecurityDescriptor()
	if err != nil {
		return fmt.Errorf("cannot initialize security descriptor for %s: %w", path, err)
	}

	if err := sd.SetDACL(acl, true, false); err != nil {
		return fmt.Errorf("cannot set DACL for %s: %w", path, err)
	}

	if err := sd.SetControl(windows.SE_DACL_PROTECTED, windows.SE_DACL_PROTECTED); err != nil {
		return fmt.Errorf("cannot protect DACL for %s: %w", path, err)
	}

	pathUTF16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("cannot encode path %s: %w", path, err)
	}

	sa := &windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: sd,
	}
	if err := windows.CreateDirectory(pathUTF16, sa); err != nil {
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			return ensureWindowsProtectedDir(path, kind)
		}

		return fmt.Errorf("cannot create %s: %w", path, err)
	}

	runtime.KeepAlive(acl)
	runtime.KeepAlive(sd)

	if err := applyWindowsProtectedDACL(path, kind); err != nil {
		return fmt.Errorf("cannot apply protected DACL to %s: %w", path, err)
	}

	return nil
}

func applyWindowsProtectedDACL(path string, kind secureDirKind) error {
	acl, err := windowsProtectedACL(kind)
	if err != nil {
		return fmt.Errorf("cannot build DACL: %w", err)
	}

	err = windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	)
	if err != nil {
		return fmt.Errorf("cannot set named security info: %w", err)
	}

	runtime.KeepAlive(acl)

	return nil
}

func windowsProtectedACL(kind secureDirKind) (*windows.ACL, error) {
	systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve SYSTEM SID: %w", err)
	}

	adminSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve Administrators SID: %w", err)
	}

	usersSID, err := windows.CreateWellKnownSid(windows.WinBuiltinUsersSid)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve Users SID: %w", err)
	}

	var pinner runtime.Pinner

	defer pinner.Unpin()

	pinner.Pin(systemSID)
	pinner.Pin(adminSID)
	pinner.Pin(usersSID)

	entries := []windows.EXPLICIT_ACCESS{
		grantWindowsSID(
			systemSID,
			windows.GENERIC_ALL,
			windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		),
		grantWindowsSID(
			adminSID,
			windows.GENERIC_ALL,
			windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		),
	}

	switch kind {
	case secureDirRoot:
		entries = append(
			entries,
			grantWindowsSID(usersSID, windows.FILE_TRAVERSE, windows.NO_INHERITANCE),
		)
	case secureDirRun:
		entries = append(
			entries,
			grantWindowsSID(
				usersSID,
				windows.FILE_GENERIC_READ|windows.FILE_GENERIC_EXECUTE,
				windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
			),
		)
	}

	acl, err := windows.ACLFromEntries(entries, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot assemble DACL: %w", err)
	}

	return acl, nil
}

func grantWindowsSID(
	sid *windows.SID,
	perm windows.ACCESS_MASK,
	inherit uint32,
) windows.EXPLICIT_ACCESS {
	return windows.EXPLICIT_ACCESS{
		AccessPermissions: perm,
		AccessMode:        windows.GRANT_ACCESS,
		Inheritance:       inherit,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_WELL_KNOWN_GROUP,
			TrusteeValue: windows.TrusteeValueFromSID(sid),
		},
	}
}

func isTrustedWindowsOwner(path string) (bool, error) {
	sd, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION,
	)
	if err != nil {
		return false, fmt.Errorf("cannot read owner: %w", err)
	}

	if sd == nil {
		return false, errors.New("security descriptor is missing")
	}

	owner, _, err := sd.Owner()
	if err != nil {
		return false, fmt.Errorf("cannot parse owner: %w", err)
	}

	if owner == nil {
		return false, errors.New("owner SID is missing")
	}

	return isTrustedWindowsSID(owner), nil
}

func isTrustedWindowsSID(sid *windows.SID) bool {
	return sid.IsWellKnown(windows.WinLocalSystemSid) ||
		sid.IsWellKnown(windows.WinBuiltinAdministratorsSid)
}

func isWindowsReparsePoint(path string) (bool, error) {
	pathUTF16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, fmt.Errorf("cannot encode path: %w", err)
	}

	attrs, err := windows.GetFileAttributes(pathUTF16)
	if err != nil {
		return false, fmt.Errorf("cannot read file attributes: %w", err)
	}

	return attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0, nil
}
