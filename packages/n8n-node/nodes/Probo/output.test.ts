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

import { describe, expect, it } from 'vitest';
import type { IDataObject } from 'n8n-workflow';

import {
	documentListFilter,
	documentVersionSelection,
	executionItems,
	formatIsoDuration,
	publishedFilterValue,
	signatureListFilter,
	stampGraphqlNode,
	waitingFields,
	withParentReference,
} from './output';

describe('executionItems', () => {
	it('returns one paired item per record', () => {
		const items = executionItems([{ id: 'a' }, { id: 'b' }], 3);
		expect(items).toEqual([
			{ json: { id: 'a' }, pairedItem: { item: 3 } },
			{ json: { id: 'b' }, pairedItem: { item: 3 } },
		]);
		expect(executionItems([], 0)).toEqual([]);
	});
});

describe('document filters', () => {
	it('shows active documents unless another status is selected', () => {
		expect(documentListFilter([], {})).toEqual({ status: ['ACTIVE'] });
		expect(documentListFilter(['ARCHIVED'], {})).toEqual({ status: ['ARCHIVED'] });
	});

	it('applies publication only when the filter is set', () => {
		expect(publishedFilterValue('any')).toBe(undefined);
		expect(publishedFilterValue(undefined)).toBe(undefined);
		expect(documentListFilter(['ACTIVE'], { published: 'any' }).published).toBe(undefined);
		expect(documentListFilter(['ACTIVE'], { published: true }).published).toBe(true);
		expect(documentListFilter(['ACTIVE'], { published: 'false' }).published).toBe(false);
		expect(documentListFilter(['ACTIVE'], { query: 'policy' }).query).toBe('policy');
	});
});

describe('signature filters', () => {
	it('keeps signature status separate from the signatory account status', () => {
		expect(signatureListFilter({ states: ['REQUESTED'], profileStates: ['ACTIVE'] })).toEqual({ states: ['REQUESTED'], profileStates: ['ACTIVE'] });
		expect(signatureListFilter({ state: ['PENDING'] })).toEqual({ profileStates: ['PENDING'] });
	});
});

describe('parent references', () => {
	it('adds documentId and versionId without dropping the nested parent', () => {
		const signature = {
			id: 'sig-1',
			documentVersion: { id: 'ver-1', document: { id: 'doc-1' } },
		};
		const stamped = withParentReference(signature);
		expect(stamped.documentId).toBe('doc-1');
		expect(stamped.versionId).toBe('ver-1');
		expect(stamped.documentVersion).toEqual(signature.documentVersion);
	});

	it('uses the known parent id when the payload does not include it', () => {
		expect(withParentReference({ id: 'ver-1' }, { documentId: 'doc-1' }).documentId).toBe('doc-1');
	});

	it('stamps the node inside a GraphQL response', () => {
		const response: IDataObject = {
			data: { node: { id: 'ver-1', document: { id: 'doc-1' } } },
		};
		const stamped = stampGraphqlNode(response, ['data', 'node'], (node) => withParentReference(node));
		const data = stamped.data as IDataObject;
		const node = data.node as IDataObject;
		expect(node.documentId).toBe('doc-1');
	});
});

describe('version selection', () => {
	it('omits the version body unless simplify is turned off', () => {
		expect(documentVersionSelection(true).includes('prosemirror')).toBe(false);
		expect(documentVersionSelection(false).includes('prosemirror')).toBe(true);
		expect(documentVersionSelection(true).includes('document')).toBe(true);
	});
});

describe('waiting duration', () => {
	it('formats the elapsed time from the moment the request started waiting', () => {
		const now = new Date('2026-10-29T15:38:00.000Z');
		expect(formatIsoDuration(0)).toBe('PT0S');
		expect(formatIsoDuration(90)).toBe('PT1M30S');
		expect(formatIsoDuration(86400)).toBe('P1D');
		expect(formatIsoDuration(90000)).toBe('P1DT1H');
		expect(waitingFields('2026-10-26T15:38:00.000Z', now)).toEqual({
			waitingSince: '2026-10-26T15:38:00.000Z',
			waitingDuration: 'P3D',
			waitingSeconds: 259200,
		});
		expect(waitingFields('not-a-date', now).waitingDuration).toBe('PT0S');
	});
});
