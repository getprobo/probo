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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['auditProgram'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Framework ID',
		name: 'frameworkId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['auditProgram'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the framework',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['auditProgram'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the audit program',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['auditProgram'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Valid From',
				name: 'validFrom',
				type: 'dateTime',
				default: '',
				description: 'The start date of the audit program validity period',
			},
			{
				displayName: 'Valid Until',
				name: 'validUntil',
				type: 'dateTime',
				default: '',
				description: 'The end date of the audit program validity period',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const frameworkId = this.getNodeParameter('frameworkId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		validFrom?: string;
		validUntil?: string;
	};

	const query = `
		mutation CreateAuditProgram($input: CreateAuditProgramInput!) {
			createAuditProgram(input: $input) {
				auditProgramEdge {
					node {
						id
						name
						validFrom
						validUntil
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		frameworkId,
		name,
	};
	if (additionalFields.validFrom) input.validFrom = additionalFields.validFrom;
	if (additionalFields.validUntil) input.validUntil = additionalFields.validUntil;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
