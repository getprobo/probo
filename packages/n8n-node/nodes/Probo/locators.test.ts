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

// The n8n package cannot depend on vitest. This file is excluded from the build.
// eslint-disable-next-line @n8n/community-nodes/no-restricted-imports
import { describe, expect, it } from 'vitest';

import { personLocator, personResultName, versionResultName } from './locators';

describe('locator labels', () => {
	it('shows a person by name and email', () => {
		expect(personResultName({
			id: 'profile-1',
			fullName: 'Ada Lovelace',
			emailAddress: 'ada@example.com',
		})).toBe('Ada Lovelace (ada@example.com)');
	});

	it('shows a document version with its number and status', () => {
		expect(versionResultName({
			id: 'version-1',
			title: 'Access policy',
			major: 1,
			minor: 2,
			status: 'PUBLISHED',
		})).toBe('Access policy v1.2 (PUBLISHED)');
	});
});

describe('personLocator', () => {
	it('keeps a distinct parameter name for each person field', () => {
		const signatory = personLocator({
			name: 'signatoryId',
			displayName: 'Signatory',
			resource: ['document'],
			operation: ['requestSignature'],
		});
		expect(signatory.name).toBe('signatoryId');
		expect(signatory.displayName).toBe('Signatory');
		expect(signatory.type).toBe('resourceLocator');
	});
});
