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

package deviceagent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrollmentLockPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	assert.Equal(
		t,
		filepath.Join(EnrollmentRunDir(dir), EnrollmentLockFileName),
		EnrollmentLockPath(dir),
	)
}

func TestAcquireEnrollmentLockConcurrent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	releaseFirst, err := AcquireEnrollmentLock(dir)
	require.NoError(t, err)

	acquired := make(chan error, 1)

	go func() {
		releaseSecond, err := AcquireEnrollmentLock(dir)
		if err != nil {
			acquired <- err
			return
		}

		releaseSecond()

		acquired <- nil
	}()

	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-acquired:
		require.NoError(t, err)
		t.Fatal("second install acquired enrollment lock while first still holds it")
	default:
	}

	releaseFirst()

	select {
	case err := <-acquired:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("second install did not acquire enrollment lock after first released it")
	}

	_, err = os.Stat(EnrollmentLockPath(dir))
	require.NoError(t, err)
}
