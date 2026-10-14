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

// Keep this list aligned with filevalidation.NewAttachmentValidator.
const attachmentTypes: ReadonlyArray<{ mimeType: string; extensions: readonly string[] }> = [
	{ mimeType: 'application/pdf', extensions: ['.pdf'] },
	{ mimeType: 'application/msword', extensions: ['.doc'] },
	{
		mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
		extensions: ['.docx'],
	},
	{ mimeType: 'application/vnd.oasis.opendocument.text', extensions: ['.odt'] },
	{ mimeType: 'application/vnd.ms-excel', extensions: ['.xls'] },
	{
		mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
		extensions: ['.xlsx'],
	},
	{ mimeType: 'application/vnd.oasis.opendocument.spreadsheet', extensions: ['.ods'] },
	{ mimeType: 'application/vnd.ms-powerpoint', extensions: ['.ppt'] },
	{
		mimeType: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
		extensions: ['.pptx'],
	},
	{ mimeType: 'application/vnd.oasis.opendocument.presentation', extensions: ['.odp'] },
	{ mimeType: 'text/markdown', extensions: ['.md'] },
	{ mimeType: 'text/plain', extensions: ['.txt', '.log'] },
	{ mimeType: 'text/x-log', extensions: ['.log'] },
	{ mimeType: 'text/uri-list', extensions: ['.uri'] },
	{ mimeType: 'text/uri-list; charset=utf-8', extensions: ['.uri'] },
	{ mimeType: 'image/jpeg', extensions: ['.jpg', '.jpeg'] },
	{ mimeType: 'image/png', extensions: ['.png'] },
	{ mimeType: 'image/svg+xml', extensions: ['.svg'] },
	{ mimeType: 'image/webp', extensions: ['.webp'] },
	{ mimeType: 'application/yaml', extensions: ['.yaml', '.yml'] },
	{ mimeType: 'application/json', extensions: ['.json'] },
	{ mimeType: 'text/yaml', extensions: ['.yaml', '.yml'] },
	{ mimeType: 'text/json', extensions: ['.json'] },
	{ mimeType: 'text/csv', extensions: ['.csv'] },
	{ mimeType: 'application/csv', extensions: ['.csv'] },
	{ mimeType: 'application/vnd.ms-excel', extensions: ['.csv'] },
];

const mimesByExtension = new Map<string, string[]>();
const extensionByMime = new Map<string, string>();

for (const type of attachmentTypes) {
	const [extension] = type.extensions;
	if (extension && !extensionByMime.has(type.mimeType)) {
		extensionByMime.set(type.mimeType, extension);
	}

	for (const candidate of type.extensions) {
		const allowed = mimesByExtension.get(candidate) ?? [];
		allowed.push(type.mimeType);
		mimesByExtension.set(candidate, allowed);
	}
}

export type AttachmentFileMetadata = {
	fileName: string;
	mimeType: string;
};

export function resolveAttachmentFile(
	fileName: string | undefined,
	mimeType: string | undefined,
): AttachmentFileMetadata {
	const name = fileName?.trim() ?? '';
	const mime = knownMime(mimeType);
	const extension = fileExtension(name);

	if (extension) {
		const allowed = mimesByExtension.get(extension);
		if (!allowed) {
			throw new Error(`File extension "${extension}" is not an allowed attachment type`);
		}

		if (!mime) {
			return { fileName: name, mimeType: allowed[0] };
		}

		if (!allowed.includes(mime)) {
			throw new Error(`MIME type "${mime}" does not match file extension "${extension}"`);
		}

		return { fileName: name, mimeType: mime };
	}

	if (!mime) {
		throw new Error('Binary data needs a supported filename or MIME type');
	}

	const inferred = extensionByMime.get(mime);
	if (!inferred) {
		throw new Error(`MIME type "${mime}" is not an allowed attachment type`);
	}

	return { fileName: `${name === '' ? 'file' : name}${inferred}`, mimeType: mime };
}

function knownMime(mimeType: string | undefined) {
	const mime = mimeType?.trim().toLowerCase() ?? '';
	if (mime === '' || mime === 'application/octet-stream') {
		return '';
	}

	return mime;
}

function fileExtension(fileName: string) {
	const slash = Math.max(fileName.lastIndexOf('/'), fileName.lastIndexOf('\\'));
	const base = slash >= 0 ? fileName.slice(slash + 1) : fileName;
	const dot = base.lastIndexOf('.');
	if (dot < 0) {
		return '';
	}

	return base.slice(dot).toLowerCase();
}
