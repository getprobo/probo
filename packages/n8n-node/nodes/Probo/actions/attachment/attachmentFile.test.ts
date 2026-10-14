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

import { resolveAttachmentFile } from './attachmentFile';

describe('resolveAttachmentFile', () => {
	it('rejects a nameless octet-stream upload', () => {
		expect(() => resolveAttachmentFile('file', 'application/octet-stream')).toThrow(
			'Binary data needs a supported filename or MIME type',
		);
		expect(() => resolveAttachmentFile(undefined, undefined)).toThrow(
			'Binary data needs a supported filename or MIME type',
		);
	});

	it('infers the MIME type from a supported filename', () => {
		expect(resolveAttachmentFile('report.PDF', '')).toEqual({
			fileName: 'report.PDF',
			mimeType: 'application/pdf',
		});
		expect(resolveAttachmentFile('data.csv', 'application/octet-stream')).toEqual({
			fileName: 'data.csv',
			mimeType: 'text/csv',
		});
	});

	it('infers a filename from a supported MIME type', () => {
		expect(resolveAttachmentFile('', 'image/png')).toEqual({
			fileName: 'file.png',
			mimeType: 'image/png',
		});
		expect(resolveAttachmentFile('scan', 'Image/JPEG')).toEqual({
			fileName: 'scan.jpg',
			mimeType: 'image/jpeg',
		});
	});

	it('keeps an explicit pair that the attachment validator allows', () => {
		expect(resolveAttachmentFile('sheet.csv', 'application/vnd.ms-excel')).toEqual({
			fileName: 'sheet.csv',
			mimeType: 'application/vnd.ms-excel',
		});
	});

	it('rejects a pair the attachment validator would refuse', () => {
		expect(() => resolveAttachmentFile('notes.exe', 'application/pdf')).toThrow(
			'File extension ".exe" is not an allowed attachment type',
		);
		expect(() => resolveAttachmentFile('photo.png', 'application/pdf')).toThrow(
			'MIME type "application/pdf" does not match file extension ".png"',
		);
		expect(() => resolveAttachmentFile(undefined, 'video/mp4')).toThrow(
			'MIME type "video/mp4" is not an allowed attachment type',
		);
	});
});
