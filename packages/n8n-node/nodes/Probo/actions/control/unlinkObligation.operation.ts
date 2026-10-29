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
import { proboApiRequest, readResourceId } from '../../GenericFunctions';
import { controlLocator, organizationLocator } from '../../locators';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['control'],
		operation: ['unlinkObligation'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
		controlLocator({
		resource: ['control'],
		operation: ['unlinkObligation'],
	}),
	{
		displayName: 'Obligation ID',
		name: 'obligationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['unlinkObligation'],
			},
		},
		default: '',
		description: 'The ID of the obligation to unlink',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const controlId = readResourceId(this.getNodeParameter('controlId', itemIndex));
	const obligationId = this.getNodeParameter('obligationId', itemIndex) as string;

	const query = `
		mutation DeleteControlObligationMapping($input: DeleteControlObligationMappingInput!) {
			deleteControlObligationMapping(input: $input) {
				deletedControlId
				deletedObligationId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { controlId, obligationId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
