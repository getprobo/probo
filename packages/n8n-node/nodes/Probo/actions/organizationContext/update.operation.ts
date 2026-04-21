// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organizationContext'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Product',
		name: 'product',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organizationContext'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The product description of the organization',
	},
	{
		displayName: 'Architecture',
		name: 'architecture',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organizationContext'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The architecture description of the organization',
	},
	{
		displayName: 'Team',
		name: 'team',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organizationContext'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The team description of the organization',
	},
	{
		displayName: 'Processes',
		name: 'processes',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organizationContext'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The processes description of the organization',
	},
	{
		displayName: 'Customers',
		name: 'customers',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organizationContext'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The customers description of the organization',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const product = this.getNodeParameter('product', itemIndex, '') as string;
	const architecture = this.getNodeParameter('architecture', itemIndex, '') as string;
	const team = this.getNodeParameter('team', itemIndex, '') as string;
	const processes = this.getNodeParameter('processes', itemIndex, '') as string;
	const customers = this.getNodeParameter('customers', itemIndex, '') as string;

	const query = `
		mutation UpdateOrganizationContext($input: UpdateOrganizationContextInput!) {
			updateOrganizationContext(input: $input) {
				context {
					organizationId
					product
					architecture
					team
					processes
					customers
				}
			}
		}
	`;

	const input: Record<string, string> = { organizationId };
	if (product) input.product = product;
	if (architecture) input.architecture = architecture;
	if (team) input.team = team;
	if (processes) input.processes = processes;
	if (customers) input.customers = customers;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
