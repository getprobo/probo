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

import {
	NodeOperationError,
	type IDataObject,
	type ILoadOptionsFunctions,
	type INodeListSearchItems,
	type INodeListSearchResult,
	type INodeProperties,
} from 'n8n-workflow';
import { proboApiRequest, proboConnectApiRequest } from './GenericFunctions';
import { readResourceId } from './output';

interface LocatorOptions {
	resource?: string[];
	operation?: string[];
	description?: string;
	required?: boolean;
}

const resourceLocatorDefault = {
	__rl: true as const,
	mode: 'list',
	value: '',
};

function withDisplayOptions(property: INodeProperties, options: LocatorOptions): INodeProperties {
	if (!options.resource || !options.operation) {
		return property;
	}

	return {
		...property,
		displayOptions: {
			show: {
				resource: options.resource,
				operation: options.operation,
			},
		},
	};
}

export function organizationLocator(options: LocatorOptions = {}): INodeProperties {
	return withDisplayOptions(
		{
			displayName: 'Organization',
			name: 'organizationId',
			type: 'resourceLocator',
			default: resourceLocatorDefault,
			required: options.required ?? true,
			description:
				options.description ?? 'The organization to use. Search the list, or enter an ID.',
			modes: [
				{
					displayName: 'From List',
					name: 'list',
					type: 'list',
					placeholder: 'Select an organization...',
					typeOptions: {
						searchListMethod: 'searchOrganizations',
						searchable: true,
					},
				},
				{
					displayName: 'By ID',
					name: 'id',
					type: 'string',
					placeholder: 'Organization ID',
					validation: [
						{
							type: 'regex',
							properties: {
								regex: '.+',
								errorMessage: 'Enter an organization ID',
							},
						},
					],
				},
			],
		},
		options,
	);
}

export function documentLocator(options: LocatorOptions = {}): INodeProperties {
	return withDisplayOptions(
		{
			displayName: 'Document',
			name: 'documentId',
			type: 'resourceLocator',
			default: resourceLocatorDefault,
			required: options.required ?? true,
			description: options.description ?? 'The document to use. Search the list, or enter an ID.',
			modes: [
				{
					displayName: 'From List',
					name: 'list',
					type: 'list',
					placeholder: 'Select a document...',
					typeOptions: {
						searchListMethod: 'searchDocuments',
						searchable: true,
					},
				},
				{
					displayName: 'By ID',
					name: 'id',
					type: 'string',
					placeholder: 'Document ID',
					validation: [
						{
							type: 'regex',
							properties: {
								regex: '.+',
								errorMessage: 'Enter a document ID',
							},
						},
					],
				},
			],
		},
		options,
	);
}

const organizationsQuery = `
	query SearchOrganizations($first: Int, $after: CursorKey, $filter: ProfileFilter) {
		viewer {
			profiles(first: $first, after: $after, filter: $filter) {
				edges {
					node {
						organization {
							id
							name
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	}
`;

const documentsQuery = `
	query SearchDocuments($organizationId: ID!, $first: Int, $after: CursorKey, $filter: DocumentFilter) {
		node(id: $organizationId) {
			... on Organization {
				documents(first: $first, after: $after, filter: $filter) {
					edges {
						node {
							id
							title
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}
`;

function matchesFilter(name: string, filter: string | undefined): boolean {
	const needle = filter?.trim().toLowerCase();
	if (!needle) {
		return true;
	}
	return name.toLowerCase().includes(needle);
}

function connectionEdges(value: unknown): IDataObject[] {
	if (!value || typeof value !== 'object' || Array.isArray(value)) {
		return [];
	}
	const edges = (value as IDataObject).edges;
	if (!Array.isArray(edges)) {
		return [];
	}
	return edges
		.map((edge) => (edge as { node?: IDataObject }).node)
		.filter((node): node is IDataObject => node !== undefined);
}

function pageInfoOf(value: unknown): { hasNextPage: boolean; endCursor?: string } {
	if (!value || typeof value !== 'object' || Array.isArray(value)) {
		return { hasNextPage: false };
	}
	const pageInfo = (value as IDataObject).pageInfo;
	if (!pageInfo || typeof pageInfo !== 'object' || Array.isArray(pageInfo)) {
		return { hasNextPage: false };
	}
	const info = pageInfo as IDataObject;
	const endCursor = typeof info.endCursor === 'string' ? info.endCursor : undefined;
	return {
		hasNextPage: Boolean(info.hasNextPage && endCursor),
		endCursor,
	};
}

export async function searchOrganizations(
	this: ILoadOptionsFunctions,
	filter?: string,
	paginationToken?: string,
): Promise<INodeListSearchResult> {
	const results: INodeListSearchItems[] = [];
	const seen = new Set<string>();
	let after = paginationToken;
	let pagination: string | undefined;
	const maxPages = filter?.trim() ? 20 : 1;

	for (let page = 0; page < maxPages; page++) {
		const response = await proboConnectApiRequest.call(this, organizationsQuery, {
			first: 100,
			after,
			filter: { states: ['ACTIVE'] },
		});
		const data = response.data as IDataObject | undefined;
		const viewer = data?.viewer as IDataObject | undefined;
		const profiles = viewer?.profiles;
		for (const membership of connectionEdges(profiles)) {
			const organization = membership.organization as IDataObject | undefined;
			const id = typeof organization?.id === 'string' ? organization.id : '';
			const name = typeof organization?.name === 'string' && organization.name ? organization.name : id;
			if (!id || seen.has(id) || !matchesFilter(name, filter)) {
				continue;
			}
			seen.add(id);
			results.push({ name, value: id });
		}

		const info = pageInfoOf(profiles);
		pagination = info.hasNextPage ? info.endCursor : undefined;
		if (!info.hasNextPage || !info.endCursor) {
			break;
		}
		if (!filter?.trim() || results.length > 0) {
			break;
		}
		after = info.endCursor;
	}

	return {
		results,
		paginationToken: pagination,
	};
}

export async function searchDocuments(
	this: ILoadOptionsFunctions,
	filter?: string,
	paginationToken?: string,
): Promise<INodeListSearchResult> {
	const organizationId = readResourceId(this.getCurrentNodeParameter('organizationId'));
	if (!organizationId) {
		throw new NodeOperationError(this.getNode(), 'Select an organization before searching documents');
	}

	const response = await proboApiRequest.call(this, documentsQuery, {
		organizationId,
		first: 50,
		after: paginationToken,
		filter: {
			...(filter?.trim() ? { query: filter.trim() } : {}),
			status: ['ACTIVE'],
		},
	});
	const data = response.data as IDataObject | undefined;
	const node = data?.node as IDataObject | undefined;
	const documents = node?.documents;
	const results = connectionEdges(documents).flatMap((document) => {
		if (typeof document.id !== 'string' || !document.id) {
			return [];
		}
		const title = typeof document.title === 'string' ? document.title : '';
		return [{ name: title || document.id, value: document.id }];
	});
	const info = pageInfoOf(documents);

	return {
		results,
		paginationToken: info.hasNextPage ? info.endCursor : undefined,
	};
}
