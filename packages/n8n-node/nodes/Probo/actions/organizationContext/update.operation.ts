// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
