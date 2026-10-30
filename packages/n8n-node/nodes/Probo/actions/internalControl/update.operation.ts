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
import { internalControlLocator, organizationLocator } from '../../locators';
import { internalControlFields, readInternalControlFields } from './internalControlFields';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['internalControl'],
		operation: ['update'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	internalControlLocator({
		name: 'id',
		resource: ['internalControl'],
		operation: ['update'],
		description: 'The internal control to update. Search the list, or enter an ID.',
	}),
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['internalControl'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the internal control',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['internalControl'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the internal control',
	},
	{
		displayName: 'Category',
		name: 'category',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['internalControl'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The category of the internal control',
	},
	{
		displayName: 'State',
		name: 'state',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['internalControl'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: 'Not Started',
				value: 'NOT_STARTED',
			},
			{
				name: 'In Progress',
				value: 'IN_PROGRESS',
			},
			{
				name: 'Not Applicable',
				value: 'NOT_APPLICABLE',
			},
			{
				name: 'Implemented',
				value: 'IMPLEMENTED',
			},
		],
		default: 'NOT_STARTED',
		description: 'The state of the internal control',
	},
	...internalControlFields('update'),
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex, '', { extractValue: true }) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const category = this.getNodeParameter('category', itemIndex, '') as string;
	const state = this.getNodeParameter('state', itemIndex, '') as string;

	const query = `
		mutation UpdateInternalControl($input: UpdateInternalControlInput!) {
			updateInternalControl(input: $input) {
				internalControl {
					id
					name
					description
					category
					code
					controlType
					nature
					operatingFrequency {
						mode
						interval
						event
					}
					evidenceCadence
					testingCadence
					nextEvidenceDue
					nextTestDue
					implementationStatus
					state
					owner {
						id
						fullName
						emailAddress
					}
					reviewer {
						id
						fullName
						emailAddress
					}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		id,
		...readInternalControlFields(
			(field) => this.getNodeParameter(field, itemIndex, '', { extractValue: true }) as string,
		),
	};
	if (name) input.name = name;
	if (description) input.description = description;
	if (category) input.category = category;
	if (state) input.state = state;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
