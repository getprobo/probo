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
	name?: string;
	displayName?: string;
	resource?: string[];
	operation?: string[];
	description?: string;
	required?: boolean;
	displayOptions?: INodeProperties['displayOptions'];
}

interface RecordLocatorSpec {
	displayName: string;
	name: string;
	searchListMethod: string;
	placeholder: string;
	idPlaceholder: string;
	idError: string;
	description: string;
}

const resourceLocatorDefault = {
	__rl: true as const,
	mode: 'list',
	value: '',
};

function withDisplayOptions(property: INodeProperties, options: LocatorOptions): INodeProperties {
	if (options.displayOptions) {
		return { ...property, displayOptions: options.displayOptions };
	}
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

function recordLocator(spec: RecordLocatorSpec, options: LocatorOptions = {}): INodeProperties {
	const displayName = options.displayName ?? spec.displayName;
	return withDisplayOptions(
		{
			displayName,
			name: options.name ?? spec.name,
			type: 'resourceLocator',
			default: resourceLocatorDefault,
			required: options.required ?? true,
			description:
				options.description ??
				`The ${displayName.toLowerCase()} to use. Search the list, or enter an ID.`,
			modes: [
				{
					displayName: 'From List',
					name: 'list',
					type: 'list',
					placeholder: spec.placeholder,
					typeOptions: {
						searchListMethod: spec.searchListMethod,
						searchable: true,
					},
				},
				{
					displayName: 'By ID',
					name: 'id',
					type: 'string',
					placeholder: spec.idPlaceholder,
					validation: [
						{
							type: 'regex',
							properties: {
								regex: '.+',
								errorMessage: spec.idError,
							},
						},
					],
				},
			],
		},
		options,
	);
}

export function personLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Person',
			name: 'ownerId',
			searchListMethod: 'searchPeople',
			placeholder: 'Select a person...',
			idPlaceholder: 'Profile ID',
			idError: 'Enter a profile ID',
			description: 'The person to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function documentVersionLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Document Version',
			name: 'documentVersionId',
			searchListMethod: 'searchDocumentVersions',
			placeholder: 'Select a version...',
			idPlaceholder: 'Document version ID',
			idError: 'Enter a document version ID',
			description: 'The document version to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function frameworkLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Framework',
			name: 'frameworkId',
			searchListMethod: 'searchFrameworks',
			placeholder: 'Select a framework...',
			idPlaceholder: 'Framework ID',
			idError: 'Enter a framework ID',
			description: 'The framework to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function controlLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Control',
			name: 'controlId',
			searchListMethod: 'searchControls',
			placeholder: 'Select a control...',
			idPlaceholder: 'Control ID',
			idError: 'Enter a control ID',
			description: 'The control to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function riskLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Risk',
			name: 'riskId',
			searchListMethod: 'searchRisks',
			placeholder: 'Select a risk...',
			idPlaceholder: 'Risk ID',
			idError: 'Enter a risk ID',
			description: 'The risk to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function internalControlLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Internal Control',
			name: 'internalControlId',
			searchListMethod: 'searchInternalControls',
			placeholder: 'Select an internal control...',
			idPlaceholder: 'Internal control ID',
			idError: 'Enter an internal control ID',
			description: 'The internal control to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function assetLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Asset',
			name: 'assetId',
			searchListMethod: 'searchAssets',
			placeholder: 'Select an asset...',
			idPlaceholder: 'Asset ID',
			idError: 'Enter an asset ID',
			description: 'The asset to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function auditLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Audit',
			name: 'auditId',
			searchListMethod: 'searchAudits',
			placeholder: 'Select an audit...',
			idPlaceholder: 'Audit ID',
			idError: 'Enter an audit ID',
			description: 'The audit to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function taskLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Task',
			name: 'taskId',
			searchListMethod: 'searchTasks',
			placeholder: 'Select a task...',
			idPlaceholder: 'Task ID',
			idError: 'Enter a task ID',
			description: 'The task to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function thirdPartyLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Third Party',
			name: 'thirdPartyId',
			searchListMethod: 'searchThirdParties',
			placeholder: 'Select a third party...',
			idPlaceholder: 'Third party ID',
			idError: 'Enter a third party ID',
			description: 'The third party to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function compliancePortalLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Compliance Portal',
			name: 'compliancePortalId',
			searchListMethod: 'searchCompliancePortals',
			placeholder: 'Select a compliance portal...',
			idPlaceholder: 'Compliance portal ID',
			idError: 'Enter a compliance portal ID',
			description: 'The compliance portal to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function cookieBannerLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Cookie Banner',
			name: 'cookieBannerId',
			searchListMethod: 'searchCookieBanners',
			placeholder: 'Select a cookie banner...',
			idPlaceholder: 'Cookie banner ID',
			idError: 'Enter a cookie banner ID',
			description: 'The cookie banner to use. Search the list, or enter an ID.',
		},
		options,
	);
}

export function accessReviewCampaignLocator(options: LocatorOptions = {}): INodeProperties {
	return recordLocator(
		{
			displayName: 'Access Review Campaign',
			name: 'accessReviewCampaignId',
			searchListMethod: 'searchAccessReviewCampaigns',
			placeholder: 'Select a campaign...',
			idPlaceholder: 'Access review campaign ID',
			idError: 'Enter an access review campaign ID',
			description: 'The access review campaign to use. Search the list, or enter an ID.',
		},
		options,
	);
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

export function personResultName(node: IDataObject): string {
	const fullName = typeof node.fullName === 'string' ? node.fullName : '';
	const email = typeof node.emailAddress === 'string' ? node.emailAddress : '';
	if (fullName && email) {
		return `${fullName} (${email})`;
	}
	return fullName || email || String(node.id ?? '');
}

export function versionResultName(node: IDataObject): string {
	const title = typeof node.title === 'string' && node.title ? node.title : 'Version';
	const major = node.major ?? '';
	const minor = node.minor ?? '';
	const status = typeof node.status === 'string' ? node.status : '';
	const version = major !== '' || minor !== '' ? ` v${major}.${minor}` : '';
	return status ? `${title}${version} (${status})` : `${title}${version}`;
}

function namedResult(node: IDataObject, fields: string[]): INodeListSearchItems | undefined {
	if (typeof node.id !== 'string' || !node.id) {
		return undefined;
	}
	for (const field of fields) {
		const value = node[field];
		if (typeof value === 'string' && value) {
			return { name: value, value: node.id };
		}
	}
	return { name: node.id, value: node.id };
}

function controlResult(node: IDataObject): INodeListSearchItems | undefined {
	const result = namedResult(node, ['name']);
	if (!result) {
		return undefined;
	}
	const section = typeof node.sectionTitle === 'string' ? node.sectionTitle : '';
	if (section && result.name !== node.id) {
		return { ...result, name: `${section} · ${result.name}` };
	}
	return result;
}

interface OrganizationSearch {
	connection: string;
	fields: string;
	filterType?: string;
	baseFilter?: IDataObject;
	label: (node: IDataObject) => INodeListSearchItems | undefined;
	filterFor?: (text: string) => IDataObject;
}

async function searchInOrganization(
	this: ILoadOptionsFunctions,
	spec: OrganizationSearch,
	filter?: string,
	paginationToken?: string,
): Promise<INodeListSearchResult> {
	const organizationId = readResourceId(this.getCurrentNodeParameter('organizationId'));
	if (!organizationId) {
		throw new NodeOperationError(this.getNode(), 'Select an organization before searching');
	}

	const filterArg = spec.filterType ? ', $filter: ' + spec.filterType : '';
	const filterField = spec.filterType ? ', filter: $filter' : '';
	const query = `
		query SearchRecords($organizationId: ID!, $first: Int, $after: CursorKey${filterArg}) {
			node(id: $organizationId) {
				... on Organization {
					${spec.connection}(first: $first, after: $after${filterField}) {
						edges {
							node {
								id
								${spec.fields}
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

	const serverFilter = Boolean(spec.filterType);
	const results: INodeListSearchItems[] = [];
	let after = paginationToken;
	let pagination: string | undefined;
	const maxPages = !serverFilter && filter?.trim() ? 20 : 1;

	for (let page = 0; page < maxPages; page++) {
		const variables: IDataObject = { organizationId, first: 50, after };
		if (serverFilter) {
			variables.filter = {
				...spec.baseFilter,
				...(filter?.trim() && spec.filterFor ? spec.filterFor(filter.trim()) : {}),
			};
		}
		const response = await proboApiRequest.call(this, query, variables);
		const data = response.data as IDataObject | undefined;
		const node = data?.node as IDataObject | undefined;
		const connection = node?.[spec.connection];
		for (const record of connectionEdges(connection)) {
			const item = spec.label(record);
			if (!item || ( !serverFilter && !matchesFilter(item.name, filter))) {
				continue;
			}
			results.push(item);
		}
		const info = pageInfoOf(connection);
		pagination = info.hasNextPage ? info.endCursor : undefined;
		if (!info.hasNextPage || !info.endCursor) {
			break;
		}
		if (serverFilter || !filter?.trim() || results.length > 0) {
			break;
		}
		after = info.endCursor;
	}

	return { results, paginationToken: pagination };
}

export async function searchPeople(
	this: ILoadOptionsFunctions,
	filter?: string,
	paginationToken?: string,
): Promise<INodeListSearchResult> {
	return searchInOrganization.call(this, {
		connection: 'profiles',
		fields: 'fullName emailAddress',
		filterType: 'ProfileFilter',
		baseFilter: { states: ['ACTIVE'] },
		filterFor: (text) => ({ query: text }),
		label: (node) => {
			if (typeof node.id !== 'string' || !node.id) {
				return undefined;
			}
			return { name: personResultName(node), value: node.id };
		},
	}, filter, paginationToken);
}

export async function searchDocumentVersions(
	this: ILoadOptionsFunctions,
	filter?: string,
	paginationToken?: string,
): Promise<INodeListSearchResult> {
	const documentId = readResourceId(this.getCurrentNodeParameter('documentId'));
	if (!documentId) {
		throw new NodeOperationError(this.getNode(), 'Select a document before searching versions');
	}

	const query = `
		query SearchDocumentVersions($documentId: ID!, $first: Int, $after: CursorKey) {
			node(id: $documentId) {
				... on Document {
					versions(first: $first, after: $after) {
						edges {
							node {
								id
								title
								major
								minor
								status
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

	const response = await proboApiRequest.call(this, query, {
		documentId,
		first: 50,
		after: paginationToken,
	});
	const data = response.data as IDataObject | undefined;
	const node = data?.node as IDataObject | undefined;
	const versions = node?.versions;
	const results = connectionEdges(versions).flatMap((version) => {
		if (typeof version.id !== 'string' || !version.id) {
			return [];
		}
		const name = versionResultName(version);
		if (!matchesFilter(name, filter)) {
			return [];
		}
		return [{ name, value: version.id }];
	});
	const info = pageInfoOf(versions);
	return {
		results,
		paginationToken: info.hasNextPage ? info.endCursor : undefined,
	};
}

function organizationSearch(spec: OrganizationSearch) {
	return function search(
		this: ILoadOptionsFunctions,
		filter?: string,
		paginationToken?: string,
	): Promise<INodeListSearchResult> {
		return searchInOrganization.call(this, spec, filter, paginationToken);
	};
}

export const searchFrameworks = organizationSearch({
	connection: 'frameworks',
	fields: 'name',
	label: (node) => namedResult(node, ['name']),
});

export const searchControls = organizationSearch({
	connection: 'controls',
	fields: 'name sectionTitle',
	filterType: 'ControlFilter',
	filterFor: (text) => ({ query: text }),
	label: controlResult,
});

export const searchRisks = organizationSearch({
	connection: 'risks',
	fields: 'name',
	filterType: 'RiskFilter',
	filterFor: (text) => ({ query: text }),
	label: (node) => namedResult(node, ['name']),
});

export const searchInternalControls = organizationSearch({
	connection: 'internalControls',
	fields: 'name code',
	filterType: 'InternalControlFilter',
	filterFor: (text) => ({ query: text }),
	label: (node) => namedResult(node, ['name', 'code']),
});

export const searchAssets = organizationSearch({
	connection: 'assets',
	fields: 'name',
	label: (node) => namedResult(node, ['name']),
});

export const searchAudits = organizationSearch({
	connection: 'audits',
	fields: 'name firm',
	label: (node) => namedResult(node, ['name', 'firm']),
});

export const searchTasks = organizationSearch({
	connection: 'tasks',
	fields: 'name',
	filterType: 'TaskFilter',
	filterFor: (text) => ({ query: text }),
	label: (node) => namedResult(node, ['name']),
});

export const searchThirdParties = organizationSearch({
	connection: 'thirdParties',
	fields: 'name',
	filterType: 'ThirdPartyFilter',
	filterFor: (text) => ({ query: text }),
	label: (node) => namedResult(node, ['name']),
});

export const searchCompliancePortals = organizationSearch({
	connection: 'compliancePortals',
	fields: 'entityName slug',
	label: (node) => namedResult(node, ['entityName', 'slug']),
});

export const searchCookieBanners = organizationSearch({
	connection: 'cookieBanners',
	fields: 'name',
	label: (node) => namedResult(node, ['name']),
});

export const searchAccessReviewCampaigns = organizationSearch({
	connection: 'accessReviewCampaigns',
	fields: 'name',
	label: (node) => namedResult(node, ['name']),
});
