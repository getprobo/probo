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

import type {
	INodeProperties,
	IExecuteFunctions,
	INodeExecutionData,
	IDataObject,
} from 'n8n-workflow';
import { proboConnectApiRequestAllItems } from '../../GenericFunctions';

const profileStates = new Set(['ACTIVE', 'DEACTIVATED', 'PENDING']);
const membershipRoles = new Set([
	'ADMIN',
	'AUDITOR',
	'COMPLIANCE_PORTAL_ACCESS_MANAGER',
	'COMPLIANCE_PORTAL_MANAGER',
	'EMPLOYEE',
	'OWNER',
	'VIEWER',
]);
const profileKinds = new Set(['CONTRACTOR', 'EMPLOYEE', 'SERVICE_ACCOUNT']);

function stringList(value: unknown): string[] {
	if (Array.isArray(value)) {
		return value.filter((item): item is string => typeof item === 'string');
	}

	if (typeof value === 'string' && value !== '') {
		return [value];
	}

	return [];
}

function allowedValues(value: unknown, allowed: Set<string>): string[] {
	return stringList(value).filter((item) => allowed.has(item));
}

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['listUsers'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['listUsers'],
			},
		},
		default: false,
		description: 'Whether to return all results or only up to a given limit',
	},
	{
		displayName: 'Limit',
		name: 'limit',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['listUsers'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of results to return',
	},
	{
		displayName: 'Search',
		name: 'query',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['listUsers'],
			},
		},
		default: '',
		description: 'Search users by full name or email address',
	},
	{
		displayName: 'States',
		name: 'state',
		type: 'multiOptions',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['listUsers'],
			},
		},
		options: [
			{ name: 'Active', value: 'ACTIVE' },
			{ name: 'Deactivated', value: 'DEACTIVATED' },
			{ name: 'Pending', value: 'PENDING' },
		],
		default: [],
		description: 'Filter by profile states',
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['listUsers'],
			},
		},
		options: [
			{ name: 'Admin', value: 'ADMIN' },
			{ name: 'All', value: '' },
			{ name: 'Auditor', value: 'AUDITOR' },
			{ name: 'Compliance Portal Access Manager', value: 'COMPLIANCE_PORTAL_ACCESS_MANAGER' },
			{ name: 'Compliance Portal Manager', value: 'COMPLIANCE_PORTAL_MANAGER' },
			{ name: 'Employee', value: 'EMPLOYEE' },
			{ name: 'Owner', value: 'OWNER' },
			{ name: 'Viewer', value: 'VIEWER' },
		],
		default: '',
		description: 'Filter by membership role',
	},
	{
		displayName: 'Type',
		name: 'kind',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['listUsers'],
			},
		},
		options: [
			{ name: 'All', value: '' },
			{ name: 'Contractor', value: 'CONTRACTOR' },
			{ name: 'Employee', value: 'EMPLOYEE' },
			{ name: 'Service Account', value: 'SERVICE_ACCOUNT' },
		],
		default: '',
		description: 'Filter by profile kind',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const query = this.getNodeParameter('query', itemIndex, '') as string;
	const states = allowedValues(this.getNodeParameter('state', itemIndex, []), profileStates);
	const [role] = allowedValues(this.getNodeParameter('role', itemIndex, ''), membershipRoles);
	const [kind] = allowedValues(this.getNodeParameter('kind', itemIndex, ''), profileKinds);

	const gqlQuery = `
		query ListUsers($organizationId: ID!, $first: Int, $after: CursorKey, $orderBy: ProfileOrder, $filter: ProfileFilter) {
			node(id: $organizationId) {
				... on Organization {
					profiles(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
						edges {
							node {
								id
								fullName
								emailAddress
								source
								state
								additionalEmailAddresses
								kind
								position
								contract {
									start
									end
								}
								createdAt
								updatedAt
								organization { id name }
								membership { id role createdAt }
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

	const filter: IDataObject = {};
	if (query) {
		filter.query = query;
	}
	if (states.length > 0) {
		filter.states = states;
	}
	if (role) {
		filter.role = role;
	}
	if (kind) {
		filter.kind = kind;
	}

	const users = await proboConnectApiRequestAllItems.call(
		this,
		gqlQuery,
		{
			organizationId,
			...(Object.keys(filter).length > 0 ? { filter } : {}),
		},
		(response: IDataObject) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			const org = node as IDataObject | undefined;
			return org?.profiles as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { users },
		pairedItem: { item: itemIndex },
	};
}
