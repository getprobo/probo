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
				resource: ['datum'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the datum',
		required: true,
	},
	{
		displayName: 'Data Classification',
		name: 'dataClassification',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Public',
				value: 'PUBLIC',
			},
			{
				name: 'Internal',
				value: 'INTERNAL',
			},
			{
				name: 'Confidential',
				value: 'CONFIDENTIAL',
			},
			{
				name: 'Secret',
				value: 'SECRET',
			},
		],
		default: 'INTERNAL',
		description: 'The classification of the data',
		required: true,
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the owner (People)',
		required: true,
	},
	{
		displayName: 'ThirdParty IDs',
		name: 'thirdPartyIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'Comma-separated list of thirdParty IDs',
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Include Owner',
				name: 'includeOwner',
				type: 'boolean',
				default: false,
				description: 'Whether to include owner details in the response',
			},
			{
				displayName: 'Include ThirdParties',
				name: 'includeThirdParties',
				type: 'boolean',
				default: false,
				description: 'Whether to include thirdParties in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const dataClassification = this.getNodeParameter('dataClassification', itemIndex) as string;
	const ownerId = this.getNodeParameter('ownerId', itemIndex) as string;
	const thirdPartyIdsStr = this.getNodeParameter('thirdPartyIds', itemIndex, '') as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeOwner?: boolean;
		includeThirdParties?: boolean;
	};

	const ownerFragment = options.includeOwner
		? `owner {
			id
			fullName
			emailAddress
		}`
		: '';

	const thirdPartiesFragment = options.includeThirdParties
		? `thirdParties(first: 100) {
			edges {
				node {
					id
					name
				}
			}
		}`
		: '';

	const query = `
		mutation CreateDatum($input: CreateDatumInput!) {
			createDatum(input: $input) {
				datumEdge {
					node {
						id
						name
						dataClassification
						${ownerFragment}
						${thirdPartiesFragment}
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const thirdPartyIds = thirdPartyIdsStr ? thirdPartyIdsStr.split(',').map((id) => id.trim()).filter(Boolean) : undefined;

	const variables = {
		input: {
			organizationId,
			name,
			dataClassification,
			ownerId,
			...(thirdPartyIds && thirdPartyIds.length > 0 && { thirdPartyIds }),
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
