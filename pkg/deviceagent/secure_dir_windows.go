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

const (
	lgIncludeIndirect  = 0x0001
	maxPreferredLength = uint32(0xFFFFFFFF)
)

var (
	modNetapi32                 = windows.NewLazySystemDLL("netapi32.dll")
	procNetApiBufferFree        = modNetapi32.NewProc("NetApiBufferFree")
	procNetLocalGroupGetMembers = modNetapi32.NewProc("NetLocalGroupGetMembers")
	procNetUserGetLocalGroups   = modNetapi32.NewProc("NetUserGetLocalGroups")
)

type (
	localGroupMembersInfo0 struct {
		sid *windows.SID
	}

	localGroupUsersInfo0 struct {
		name *uint16
	}
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

	if err := applyWindowsProtectedTree(path, kind); err != nil {
		return fmt.Errorf("cannot apply protected DACL to %s: %w", path, err)
	}

	return nil
}

func createWindowsProtectedDir(path string, kind secureDirKind) error {
	err := createWindowsDirectory(path, kind, true)
	if err != nil {
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			return ensureWindowsProtectedDir(path, kind)
		}

		err = createWindowsDirectory(path, kind, false)
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			return ensureWindowsProtectedDir(path, kind)
		}

		if err != nil {
			return err
		}
	}

	if err := applyWindowsProtectedTree(path, kind); err != nil {
		return fmt.Errorf("cannot apply protected DACL to %s: %w", path, err)
	}

	return nil
}

func createWindowsDirectory(path string, kind secureDirKind, setAdminOwner bool) error {
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

	var adminSID *windows.SID

	if setAdminOwner {
		adminSID, err = windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
		if err != nil {
			return fmt.Errorf("cannot resolve Administrators SID: %w", err)
		}

		if err := sd.SetOwner(adminSID, false); err != nil {
			return fmt.Errorf("cannot set owner for %s: %w", path, err)
		}
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
		return fmt.Errorf("cannot create %s: %w", path, err)
	}

	runtime.KeepAlive(acl)
	runtime.KeepAlive(sd)
	runtime.KeepAlive(adminSID)

	return nil
}

func applyWindowsProtectedTree(path string, kind secureDirKind) error {
	if err := applyWindowsProtectedDACL(path, kind); err != nil {
		return err
	}

	if kind != secureDirAgent {
		return nil
	}

	return applyWindowsProtectedDACLChildren(path, kind)
}

func applyWindowsProtectedDACLChildren(dir string, kind secureDirKind) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot list %s: %w", dir, err)
	}

	for _, entry := range entries {
		child := filepath.Join(dir, entry.Name())

		reparse, err := isWindowsReparsePoint(child)
		if err != nil {
			return fmt.Errorf("cannot inspect %s: %w", child, err)
		}

		if reparse {
			return fmt.Errorf("%s is a reparse point; remove it and retry", child)
		}

		if err := applyWindowsProtectedDACL(child, kind); err != nil {
			return fmt.Errorf("cannot apply protected DACL to %s: %w", child, err)
		}

		if entry.IsDir() {
			if err := applyWindowsProtectedDACLChildren(child, kind); err != nil {
				return err
			}
		}
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
	if sid == nil {
		return false
	}

	if sid.IsWellKnown(windows.WinLocalSystemSid) ||
		sid.IsWellKnown(windows.WinBuiltinAdministratorsSid) {
		return true
	}

	return sidIsLocalAdministrator(sid)
}

func sidIsLocalAdministrator(sid *windows.SID) bool {
	return localAdministratorsContainsSID(sid) || localAdministratorsContainsUser(sid)
}

func builtinAdministratorsName() (string, error) {
	adminSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return "", fmt.Errorf("cannot resolve Administrators SID: %w", err)
	}

	name, _, _, err := adminSID.LookupAccount("")
	if err != nil {
		return "", fmt.Errorf("cannot look up Administrators name: %w", err)
	}

	return name, nil
}

func localAdministratorsContainsSID(sid *windows.SID) bool {
	name, err := builtinAdministratorsName()
	if err != nil {
		return false
	}

	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return false
	}

	var (
		buf          *localGroupMembersInfo0
		entriesRead  uint32
		totalEntries uint32
		resumeHandle uintptr
	)

	status, _, _ := procNetLocalGroupGetMembers.Call(
		0,
		uintptr(unsafe.Pointer(namePtr)),
		0,
		uintptr(unsafe.Pointer(&buf)),
		uintptr(maxPreferredLength),
		uintptr(unsafe.Pointer(&entriesRead)),
		uintptr(unsafe.Pointer(&totalEntries)),
		uintptr(unsafe.Pointer(&resumeHandle)),
	)
	if status != 0 || buf == nil || entriesRead == 0 {
		return false
	}

	defer procNetApiBufferFree.Call(uintptr(unsafe.Pointer(buf)))

	members := unsafe.Slice(buf, entriesRead)
	for _, member := range members {
		if member.sid != nil && member.sid.Equals(sid) {
			return true
		}
	}

	return false
}

func localAdministratorsContainsUser(sid *windows.SID) bool {
	account, _, _, err := sid.LookupAccount("")
	if err != nil || account == "" {
		return false
	}

	accountPtr, err := windows.UTF16PtrFromString(account)
	if err != nil {
		return false
	}

	adminName, err := builtinAdministratorsName()
	if err != nil {
		return false
	}

	var (
		buf          *localGroupUsersInfo0
		entriesRead  uint32
		totalEntries uint32
	)

	status, _, _ := procNetUserGetLocalGroups.Call(
		0,
		uintptr(unsafe.Pointer(accountPtr)),
		0,
		lgIncludeIndirect,
		uintptr(unsafe.Pointer(&buf)),
		uintptr(maxPreferredLength),
		uintptr(unsafe.Pointer(&entriesRead)),
		uintptr(unsafe.Pointer(&totalEntries)),
	)
	if status != 0 || buf == nil || entriesRead == 0 {
		return false
	}

	defer procNetApiBufferFree.Call(uintptr(unsafe.Pointer(buf)))

	groups := unsafe.Slice(buf, entriesRead)
	for _, group := range groups {
		if group.name == nil {
			continue
		}

		if strings.EqualFold(windows.UTF16PtrToString(group.name), adminName) {
			return true
		}
	}

	return false
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
