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

package cookiebanner

import "errors"

var (
	ErrBannerNotFound             = errors.New("cookie banner not found")
	ErrCategoryNotFound           = errors.New("cookie category not found")
	ErrVersionNotFound            = errors.New("cookie banner version not found")
	ErrBannerAlreadyActive        = errors.New("cookie banner is already active")
	ErrBannerAlreadyInactive      = errors.New("cookie banner is already inactive")
	ErrVersionNotPublished        = errors.New("cookie banner version is not published")
	ErrNoPublishedVersion         = errors.New("no published cookie banner version")
	ErrNoDraftVersion             = errors.New("no draft cookie banner version to publish")
	ErrCannotDeleteSystemCategory = errors.New("cannot delete system cookie category")
	ErrCategorySlugAlreadyExists  = errors.New("a category with this slug already exists in this banner")
	ErrOriginAlreadyInUse         = errors.New("origin is already used by another active cookie banner")
	ErrConsentNotFound            = errors.New("consent record not found")
	ErrCookieNotFound             = errors.New("cookie not found")
	ErrCategoriesBannerMismatch   = errors.New("source and target categories belong to different banners")
	ErrPostHogConsentKindInvalid  = errors.New("PostHog consent can only be enabled on normal categories")
	ErrTrackerPatternNotFound     = errors.New("tracker pattern not found")
	ErrPatternAlreadyExists       = errors.New("a pattern with this name already exists in this banner")
	ErrSamePatternCategoryMove    = errors.New("source and target cookie categories must be different")
	ErrTrackerResourceNotFound    = errors.New("tracker resource not found")
	ErrResourceAlreadyExists      = errors.New("a resource with this origin and path already exists in this banner")
	ErrSameResourceCategoryMove   = errors.New("source and target cookie categories must be different")
)
