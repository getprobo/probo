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

import type { IDataObject, INodeExecutionData } from 'n8n-workflow';

export const profileSelection = `id
fullName
emailAddress`;

export function executionItems(records: IDataObject[], itemIndex: number): INodeExecutionData[] {
	return records.map((json) => ({
		json,
		pairedItem: { item: itemIndex },
	}));
}

export function stringArray(value: unknown): string[] {
	if (!Array.isArray(value)) {
		return [];
	}

	return value.filter((item): item is string => typeof item === 'string' && item.length > 0);
}

export function publishedFilterValue(value: unknown): boolean | undefined {
	if (value === true || value === 'true') {
		return true;
	}
	if (value === false || value === 'false') {
		return false;
	}
	return undefined;
}

export function documentListFilter(status: unknown, filters: IDataObject): IDataObject {
	const filter: IDataObject = {};
	if (typeof filters.query === 'string' && filters.query) {
		filter.query = filters.query;
	}

	const writeModes = stringArray(filters.writeModes);
	if (writeModes.length) {
		filter.writeModes = writeModes;
	}

	const documentTypes = stringArray(filters.documentTypes);
	if (documentTypes.length) {
		filter.documentTypes = documentTypes;
	}

	const classifications = stringArray(filters.classifications);
	if (classifications.length) {
		filter.classifications = classifications;
	}

	const published = publishedFilterValue(filters.published);
	if (published !== undefined) {
		filter.published = published;
	}

	const selected = stringArray(status);
	filter.status = selected.length ? selected : ['ACTIVE'];
	return filter;
}

export function signatureListFilter(filters: IDataObject): IDataObject {
	const filter: IDataObject = {};
	const states = stringArray(filters.states);
	if (states.length) {
		filter.states = states;
	}
	if (filters.activeContract !== undefined) {
		filter.activeContract = filters.activeContract;
	}

	// Saved workflows named the signatory account filter `state`.
	const signatoryStatus = stringArray(filters.profileStates);
	const legacyStatus = stringArray(filters.state);
	const profileStates = signatoryStatus.length ? signatoryStatus : legacyStatus;
	if (profileStates.length) {
		filter.profileStates = profileStates;
	}

	return filter;
}

function nestedObject(value: unknown): IDataObject | undefined {
	if (value && typeof value === 'object' && !Array.isArray(value)) {
		return value as IDataObject;
	}
	return undefined;
}

function stringId(value: unknown): string | undefined {
	return typeof value === 'string' && value.length > 0 ? value : undefined;
}

export function withParentReference(
	record: IDataObject,
	fallback: { documentId?: string; versionId?: string } = {},
): IDataObject {
	const version = nestedObject(record.documentVersion);
	const document = nestedObject(version?.document) ?? nestedObject(record.document);
	const documentId = stringId(document?.id) ?? fallback.documentId;
	const versionId = stringId(version?.id) ?? fallback.versionId;

	return {
		...record,
		...(documentId ? { documentId } : {}),
		...(versionId ? { versionId } : {}),
	};
}

export function stampGraphqlNode(
	response: IDataObject,
	path: string[],
	stamp: (node: IDataObject) => IDataObject,
): IDataObject {
	const current = path.reduce<unknown>((value, key) => {
		if (!value || typeof value !== 'object' || Array.isArray(value)) {
			return undefined;
		}
		return (value as IDataObject)[key];
	}, response);

	if (!current || typeof current !== 'object' || Array.isArray(current)) {
		return response;
	}

	return setAt(response, path, stamp(current as IDataObject));
}

function setAt(root: IDataObject, path: string[], next: IDataObject): IDataObject {
	if (path.length === 0) {
		return next;
	}

	const [head, ...rest] = path;
	const child = root[head];
	const updated =
		child && typeof child === 'object' && !Array.isArray(child)
			? setAt(child as IDataObject, rest, next)
			: next;

	return { ...root, [head]: updated };
}

export function documentVersionSelection(simplify: boolean): string {
	const content = simplify
		? ''
		: `content {
								prosemirror
							}`;

	return `id
							title
							major
							minor
							status
							${content}
							changelog
							classification
							documentType
							publishedAt
							createdAt
							updatedAt
							document {
								id
							}`;
}

export function personFields(person: IDataObject | undefined): IDataObject {
	return {
		id: person?.id ?? null,
		fullName: person?.fullName ?? null,
		emailAddress: person?.emailAddress ?? null,
	};
}

export function formatIsoDuration(totalSeconds: number): string {
	const seconds = Math.max(0, Math.floor(totalSeconds));
	const days = Math.floor(seconds / 86400);
	const hours = Math.floor((seconds % 86400) / 3600);
	const minutes = Math.floor((seconds % 3600) / 60);
	const remain = seconds % 60;

	let duration = 'P';
	if (days > 0) {
		duration += `${days}D`;
	}
	if (hours > 0 || minutes > 0 || remain > 0 || days === 0) {
		duration += 'T';
		if (hours > 0) {
			duration += `${hours}H`;
		}
		if (minutes > 0) {
			duration += `${minutes}M`;
		}
		if (remain > 0 || (hours === 0 && minutes === 0)) {
			duration += `${remain}S`;
		}
	}

	return duration;
}

export function waitingFields(since: unknown, now: Date): IDataObject {
	const waitingSince = typeof since === 'string' ? since : '';
	const start = Date.parse(waitingSince);
	const waitingSeconds = Number.isNaN(start)
		? 0
		: Math.max(0, Math.floor((now.getTime() - start) / 1000));

	return {
		waitingSince: waitingSince || null,
		waitingDuration: formatIsoDuration(waitingSeconds),
		waitingSeconds,
	};
}
